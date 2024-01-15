package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/handlers"
	myHandler "github.com/gorilla/handlers" // Importa el paquete handlers
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var store *sessions.CookieStore

func main() {
	// Inicializar la conexión a la base de datos
	db.InitDB()
	defer db.CloseDB()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error al cargar el archivo .env:", err)
	}

	sessionKey := []byte(os.Getenv("SESSION_KEY"))

	store = sessions.NewCookieStore(sessionKey)

	router := mux.NewRouter()

	// Ruta para servir archivos estáticos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Rutas
	router.HandleFunc("/", handlers.HomeHandler)
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/signup", handlers.SignupHandler)
	router.HandleFunc("/dashboard", handlers.DashboardHandler)
	router.HandleFunc("/logout", handlers.LogoutHandler)

	// rutas de post en api
	router.HandleFunc("/api/posts", handlers.GetAllPostsHandler).Methods("GET")
	router.HandleFunc("/api/create-post", handlers.CreatePostHandler).Methods("POST")
	router.HandleFunc("/api/update-post/{postID}", handlers.UpdatePostHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/api/delete-post/{postID}", handlers.DeletePostHandler).Methods("DELETE")

	// Configuracion el middleware CORS
	corsHandler := myHandler.CORS(
		myHandler.AllowedOrigins([]string{"*"}), // Permite solicitudes desde cualquier origen
		myHandler.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		myHandler.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	addr := ":8080"
	fmt.Printf("Servidor corriendo en http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, corsHandler(router))) // Utiliza el middleware CORS con el enrutador principal
}
