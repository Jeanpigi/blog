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
		log.Println("⚠️ No se encontró el archivo .env, usando variables de entorno del sistema.")
	}

	// 🔹 Obtener la clave de sesión
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("❌ Error: SESSION_KEY no está definida. Verifica tus variables de entorno.")
	}

	// 🔹 Inicializar sesión (sin argumentos)
	session.InitStore()

	// 🔹 Configuración del router
	router := mux.NewRouter()

	// 🔹 Servir archivos estáticos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// 🔹 Middleware para registrar visitas en rutas específicas
	router.Handle("/", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.HomeHandler))).Methods("GET")
	router.Handle("/blog", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.BlogHandler))).Methods("GET")
	router.Handle("/post/{id}", middleware.TrackVisitMiddleware(http.HandlerFunc(handlers.PostHandler))).Methods("GET")

	// 🔹 Rutas de autenticación y dashboard
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/signup", handlers.SignupHandler)
	router.HandleFunc("/dashboard", handlers.DashboardHandler)
	router.HandleFunc("/logout", handlers.LogoutHandler)

	// 🔹 Otras rutas de contenido
	router.HandleFunc("/portafolio", handlers.PortafolioHandler)
	router.HandleFunc("/historias", handlers.HistoriasHandler)
	router.HandleFunc("/tecnologias", handlers.TecnologiasHandler)
	router.HandleFunc("/visitas", handlers.VisitsPageHandler).Methods("GET")

	// 🔹 Rutas API para posts
	router.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods("GET")
	router.HandleFunc("/api/posts/{id}", handlers.GetPostsHandler).Methods("GET")
	router.HandleFunc("/api/create-post", handlers.CreatePostHandler).Methods("POST")
	router.HandleFunc("/api/update-post/{postID}", handlers.UpdatePostHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/api/delete-post/{postID}", handlers.DeletePostHandler).Methods("DELETE")

	// 🔹 Rutas API para categorías e historias
	router.HandleFunc("/api/categories", handlers.GetPostsByCategoryHandler).Methods("GET")
	router.HandleFunc("/api/histories", handlers.GetPostsByHistoryHandler).Methods("GET")

	// 🔹 Rutas API para visitas
	router.HandleFunc("/api/visits/location", handlers.GetVisitsWithLocationHandler).Methods("GET")

	// 🔹 Configuración del middleware CORS
	corsHandler := myHandler.CORS(
		myHandler.AllowedOrigins([]string{"*"}), // Permite solicitudes desde cualquier origen
		myHandler.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		myHandler.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// 🔹 Iniciar el servidor
	addr := ":8080"
	fmt.Printf("🚀 Servidor corriendo en http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, corsHandler(router)); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}

