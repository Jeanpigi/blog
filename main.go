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

	// 🔹 Configurar router principal
	router := mux.NewRouter()

	// 🔹 Servir archivos estáticos compartidos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// 🔹 Middleware de visitas para rutas del blog
	router.Handle("/", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.HomeHandler))).Methods("GET", "HEAD")
	router.Handle("/blog", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.BlogHandler))).Methods("GET", "HEAD")
	router.Handle("/post/{id}", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.PostHandler))).Methods("GET", "HEAD")

	// 🔹 Rutas de autenticación y dashboard
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/signup", handlers.SignupHandler)
	router.HandleFunc("/dashboard", middleware.RequireAuth(handlers.DashboardHandler)).Methods("GET", "HEAD")
	router.HandleFunc("/logout", handlers.LogoutHandler)

	// 🔹 Rutas de contenido adicional
	router.HandleFunc("/portafolio", handlers.PortafolioHandler)
	router.HandleFunc("/historias", handlers.HistoriasHandler)
	router.HandleFunc("/tecnologias", handlers.TecnologiasHandler)
	router.HandleFunc("/visitas", handlers.VisitsPageHandler).Methods("GET", "HEAD")

	// 🔹 API: Posts
	router.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods("GET", "HEAD")
	router.HandleFunc("/api/posts/{id}", handlers.GetPostsHandler).Methods("GET", "HEAD")
	router.HandleFunc("/api/create-post", middleware.RequireAuth(middleware.RequireCSRF(handlers.CreatePostHandler))).Methods("POST")
	router.HandleFunc("/api/update-post/{postID}", middleware.RequireAuth(middleware.RequireCSRF(handlers.UpdatePostHandler))).Methods("PUT", "PATCH")
	router.HandleFunc("/api/delete-post/{postID}", middleware.RequireAuth(middleware.RequireCSRF(handlers.DeletePostHandler))).Methods("DELETE")

	// 🔹 API: Categorías e historias
	router.HandleFunc("/api/categories", handlers.GetPostsByCategoryHandler).Methods("GET", "HEAD")
	router.HandleFunc("/api/histories", handlers.GetPostsByHistoryHandler).Methods("GET", "HEAD")

	// 🔹 API: Visitas (expone IPs y geolocalización de visitantes: solo el admin)
	router.HandleFunc("/api/visits/location", middleware.RequireAuth(handlers.GetVisitsWithLocationHandler)).Methods("GET", "HEAD")

	// ✅ RUTAS DE RADIO (integradas)
	router.HandleFunc("/radio/stream", handlers.StreamHandler).Methods("GET", "HEAD")
	router.HandleFunc("/radio/upload", middleware.RequireAuth(handlers.UploadHandler)).Methods("GET", "POST")
	router.HandleFunc("/api/radio/now-playing", handlers.NowPlayingHandler).Methods("GET", "HEAD")
	// Saltar canción es una acción de admin (afecta a todos los oyentes del broadcast global).
	router.HandleFunc("/api/radio/advance", middleware.RequireAuth(handlers.AdvanceSongHandler)).Methods("POST")

	// 🔹 Handler para rutas inexistentes
	router.NotFoundHandler = http.HandlerFunc(handlers.NotFoundHandler)

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
	addr := ":8080"
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


