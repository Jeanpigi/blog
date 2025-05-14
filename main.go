package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	// 🔹 Inicializar la conexión a la base de datos
	db.InitDB()
	defer db.CloseDB()

	// 🔹 Cargar variables de entorno desde `.env` si existe
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ No se pudo cargar el archivo .env, se usarán las variables del sistema.")
		}
	} else {
		log.Println("⚠️ No se encontró el archivo .env, usando variables del sistema.")
	}

	// 🔹 Verificar clave de sesión
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("❌ Error: SESSION_KEY no está definida. Verifica tus variables de entorno.")
	}

	// 🔹 Inicializar sesión
	session.InitStore()

	// 🔹 Inicializar música y playlist
	musicFolder := "./music"
	if err := music.LoadMusicFiles(musicFolder); err != nil {
		log.Fatalf("❌ Error al cargar archivos de música: %v", err)
	}
	playlist.CreatePlaylist()

	// 🔹 Configurar router principal
	router := mux.NewRouter()

	// 🔹 Servir archivos estáticos compartidos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// 🔹 Middleware de visitas para rutas del blog
	router.Handle("/", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.HomeHandler))).Methods("GET")
	router.Handle("/blog", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.BlogHandler))).Methods("GET")
	router.Handle("/post/{id}", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.PostHandler))).Methods("GET")

	// 🔹 Rutas de autenticación y dashboard
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/signup", handlers.SignupHandler)
	router.HandleFunc("/dashboard", middleware.RequireAuth(handlers.DashboardHandler)).Methods("GET")
	router.HandleFunc("/logout", handlers.LogoutHandler)

	// 🔹 Rutas de contenido adicional
	router.HandleFunc("/portafolio", handlers.PortafolioHandler)
	router.HandleFunc("/historias", handlers.HistoriasHandler)
	router.HandleFunc("/tecnologias", handlers.TecnologiasHandler)
	router.HandleFunc("/visitas", handlers.VisitsPageHandler).Methods("GET")

	// 🔹 API: Posts
	router.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods("GET")
	router.HandleFunc("/api/posts/{id}", handlers.GetPostsHandler).Methods("GET")
	router.HandleFunc("/api/create-post", middleware.RequireAuth(handlers.CreatePostHandler)).Methods("POST")
	router.HandleFunc("/api/update-post/{postID}", middleware.RequireAuth(handlers.UpdatePostHandler)).Methods("PUT", "PATCH")
	router.HandleFunc("/api/delete-post/{postID}", middleware.RequireAuth(handlers.DeletePostHandler)).Methods("DELETE")

	// 🔹 API: Categorías e historias
	router.HandleFunc("/api/categories", handlers.GetPostsByCategoryHandler).Methods("GET")
	router.HandleFunc("/api/histories", handlers.GetPostsByHistoryHandler).Methods("GET")

	// 🔹 API: Visitas
	router.HandleFunc("/api/visits/location", handlers.GetVisitsWithLocationHandler).Methods("GET")

	// ✅ RUTAS DE RADIO (integradas)
	router.HandleFunc("/radio/stream", handlers.StreamHandler).Methods("GET")
	router.HandleFunc("/radio/upload", middleware.RequireAuth(handlers.UploadHandler)).Methods("GET", "POST")

	// 🔹 Middleware CORS
	corsHandler := myHandler.CORS(
		myHandler.AllowedOrigins([]string{"*"}),
		myHandler.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		myHandler.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// 🔹 Iniciar servidor
	addr := ":8080"
	fmt.Printf("🚀 Servidor unificado corriendo en http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, corsHandler(router)); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}


