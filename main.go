package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/handlers"
	"github.com/Jeanpigi/blog/session"
	myHandler "github.com/gorilla/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// Inicializar la conexión a la base de datos
	db.InitDB()
	defer db.CloseDB()

	// Verificar si existe un archivo .env antes de intentar cargarlo
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ No se pudo cargar el archivo .env, se usarán las variables del sistema.")
		}
	} else {
		log.Println("⚠️ No se encontró el archivo .env, usando variables de entorno del sistema.")
	}

	// Obtener la clave de sesión (ya sea desde .env o el sistema)
	sessionKey := []byte(os.Getenv("SESSION_KEY"))
	if len(sessionKey) == 0 {
		log.Fatal("❌ Error: SESSION_KEY no está definida. Verifica tus variables de entorno.")
	}
	session.InitStore(sessionKey)

	router := mux.NewRouter()

	// Ruta para servir archivos estáticos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Rutas
	router.HandleFunc("/", handlers.HomeHandler)
	router.HandleFunc("/post/{id}", handlers.PostHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/signup", handlers.SignupHandler)
	router.HandleFunc("/dashboard", handlers.DashboardHandler)
	router.HandleFunc("/logout", handlers.LogoutHandler)
	router.HandleFunc("/portafolio", handlers.PortafolioHandler)
	router.HandleFunc("/blog", handlers.BlogHandler)
	router.HandleFunc("/historias", handlers.HistoriasHandler)
	router.HandleFunc("/tecnologias", handlers.TecnologiasHandler)

	// rutas de post en api
	router.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods("GET")
	router.HandleFunc("/api/posts/{id}", handlers.GetPostsHandler).Methods("GET")
	router.HandleFunc("/api/create-post", handlers.CreatePostHandler).Methods("POST")
	router.HandleFunc("/api/update-post/{postID}", handlers.UpdatePostHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/api/delete-post/{postID}", handlers.DeletePostHandler).Methods("DELETE")

	// rutas de categorias e historias en api
	router.HandleFunc("/api/categories", handlers.GetPostsByCategoryHandler).Methods("GET")
	router.HandleFunc("/api/histories", handlers.GetPostsByHistoryHandler).Methods("GET")

	// Configuracion del middleware CORS
	corsHandler := myHandler.CORS(
		myHandler.AllowedOrigins([]string{"*"}), // Permite solicitudes desde cualquier origen
		myHandler.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		myHandler.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	addr := ":8080"
	fmt.Printf("🚀 Servidor corriendo en http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, corsHandler(router))) // Utiliza el middleware CORS con el enrutador principal
}
