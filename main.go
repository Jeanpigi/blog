package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/handlers"
	"github.com/Jeanpigi/blog/internal/middleware"
	"github.com/Jeanpigi/blog/internal/music"
	"github.com/Jeanpigi/blog/internal/playlist"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
	myHandler "github.com/gorilla/handlers"
)

func main() {
	// 🔹 Cargar variables de entorno PRIMERO (antes de InitDB)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No se pudo cargar el archivo .env, se usarán las variables del sistema.")
		}
	} else {
		log.Println("⚠️ No se encontró el archivo .env, usando variables del sistema.")
	}

	// 🔹 Inicializar la conexión a la base de datos
	db.InitDB()
	defer db.CloseDB()

	// 🔹 Verificar claves de sesión
	if os.Getenv("SESSION_AUTH_KEY") == "" || os.Getenv("SESSION_ENC_KEY") == "" {
		log.Fatal("❌ Error: SESSION_AUTH_KEY y SESSION_ENC_KEY deben estar definidas.")
	}

	// 🔹 Inicializar sesión
	session.InitStore()

	// 🔹 Inicializar música y playlist
	musicFolder := "./music"
	if err := os.MkdirAll(musicFolder, 0755); err != nil {
		log.Printf("⚠️ No se pudo crear la carpeta de música: %v", err)
	}
	if err := music.LoadMusicFiles(musicFolder); err != nil {
		log.Printf("⚠️ Error al cargar archivos de música: %v", err)
	}
	playlist.CreatePlaylist()
	handlers.InitBroadcast()

	// 🔹 Limpieza diaria de visitas (bots, IPs reales, retención)
	handlers.StartVisitsMaintenance()

	// 🔹 Versión de assets para cache-busting (?v=...). En producción conviene
	// fijarla al hash de git: STATIC_VERSION=$(git rev-parse --short HEAD)
	utils.SetStaticVersion(os.Getenv("STATIC_VERSION"))

	router := newRouter()

	// 🔹 Middleware CORS
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
		log.Println("⚠️ ALLOWED_ORIGIN no está definido: CORS permitirá cualquier origen (*). Defínelo en producción.")
	}
	corsHandler := myHandler.CORS(
		myHandler.AllowedOrigins([]string{allowedOrigin}),
		myHandler.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		myHandler.AllowedHeaders([]string{"Content-Type", "Authorization", "X-CSRF-Token"}),
	)

	// 🔹 Cadena de middleware global: security headers → CORS → router
	rootHandler := middleware.SecurityHeaders(corsHandler(router))

	// 🔹 Configurar servidor con timeouts (defensa contra Slowloris).
	// WriteTimeout se deja en 0 a propósito: el stream de la radio es una
	// descarga larga y un WriteTimeout global la cortaría. nginx/Cloudflare
	// aplican sus propios límites de escritura por delante.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	srv := &http.Server{
		Addr:              addr,
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 🔹 Iniciar servidor en goroutine para permitir apagado ordenado
	go func() {
		fmt.Printf("🚀 Servidor unificado corriendo en http://localhost%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error al iniciar el servidor: %v", err)
		}
	}()

	// 🔹 Apagado ordenado: espera SIGINT/SIGTERM (deploys, systemd restart)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("🛑 Apagando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ Error durante el apagado: %v", err)
	}
	log.Println("✅ Servidor detenido limpiamente.")
}

// newRouter registra todas las rutas. Convenciones:
//   - Páginas HTML: GET/HEAD. Las de admin redirigen a /login sin sesión.
//   - API pública: GET /api/... (JSON).
//   - API admin: auth → 401 JSON; mutaciones además exigen X-CSRF-Token.
//   - Posts siguen REST: /api/posts y /api/posts/{id} con el verbo indicando la acción.
func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true) // /blog/ → /blog

	auth := middleware.RequireAuth
	authCSRF := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.RequireAuth(middleware.RequireCSRF(h))
	}
	page := []string{http.MethodGet, http.MethodHead}

	// 🔹 Estáticos y healthcheck
	r.PathPrefix("/static/").Handler(middleware.StaticCache(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))))
	r.HandleFunc("/healthz", handlers.HealthHandler).Methods(page...)

	// 🔹 Páginas públicas (las visitas las mide static/js/track.js → /api/track)
	r.HandleFunc("/", handlers.HomeHandler).Methods(page...)
	r.HandleFunc("/blog", handlers.BlogHandler).Methods(page...)
	r.HandleFunc("/post/{id:[0-9]+}", handlers.PostHandler).Methods(page...)
	r.HandleFunc("/historias", handlers.HistoriasHandler).Methods(page...)
	r.HandleFunc("/tecnologias", handlers.TecnologiasHandler).Methods(page...)
	r.HandleFunc("/portafolio", handlers.PortafolioHandler).Methods(page...)

	// 🔹 Autenticación
	r.HandleFunc("/login", handlers.LoginHandler).Methods(http.MethodGet, http.MethodHead, http.MethodPost)
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods(http.MethodPost)
	r.HandleFunc("/signup", auth(handlers.SignupHandler)).Methods(http.MethodGet, http.MethodHead, http.MethodPost)

	// 🔹 Páginas de admin
	r.HandleFunc("/dashboard", auth(handlers.DashboardHandler)).Methods(page...)
	r.HandleFunc("/visitas", auth(handlers.VisitsPageHandler)).Methods(page...) // expone IPs/geo
	r.HandleFunc("/radio/upload", auth(handlers.UploadHandler)).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/radio/upload", authCSRF(handlers.UploadHandler)).Methods(http.MethodPost)

	// 🔹 API pública
	r.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods(page...)
	r.HandleFunc("/api/posts/{id:[0-9]+}", handlers.GetPostsHandler).Methods(page...)
	r.HandleFunc("/api/categories", handlers.GetPostsByCategoryHandler).Methods(page...)
	r.HandleFunc("/api/histories", handlers.GetPostsByHistoryHandler).Methods(page...)
	r.HandleFunc("/api/radio/now-playing", handlers.NowPlayingHandler).Methods(page...)
	r.HandleFunc("/radio/stream", handlers.StreamHandler).Methods(page...)
	// Medición de visitas desde el navegador (sendBeacon no envía cabeceras
	// CSRF; son endpoints sin estado sensible, validados y con límite por IP).
	r.HandleFunc("/api/track", handlers.TrackStartHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/track/end", handlers.TrackEndHandler).Methods(http.MethodPost)

	// 🔹 API admin
	r.HandleFunc("/api/posts", authCSRF(handlers.CreatePostHandler)).Methods(http.MethodPost)
	r.HandleFunc("/api/posts/{id:[0-9]+}", authCSRF(handlers.UpdatePostHandler)).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/api/posts/{id:[0-9]+}", authCSRF(handlers.DeletePostHandler)).Methods(http.MethodDelete)
	r.HandleFunc("/api/admin/stats", auth(handlers.DashboardStatsHandler)).Methods(page...)
	r.HandleFunc("/api/admin/analytics", auth(handlers.AnalyticsHandler)).Methods(page...)
	r.HandleFunc("/api/visits/location", auth(handlers.GetVisitsWithLocationHandler)).Methods(page...)
	// Saltar canción afecta a todos los oyentes del broadcast global.
	r.HandleFunc("/api/radio/advance", authCSRF(handlers.AdvanceSongHandler)).Methods(http.MethodPost)

	r.NotFoundHandler = http.HandlerFunc(handlers.NotFoundHandler)
	return r
}
