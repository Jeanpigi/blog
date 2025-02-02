package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

// SignupHandler protege la ruta para que solo usuarios autenticados puedan acceder
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar si el usuario está autenticado
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Si la petición es GET, renderizar el formulario de registro
	if r.Method == "GET" {
		utils.RenderTemplate(w, "templates/signup.html", nil)
		return
	}

	// Si la petición es POST, procesar el formulario de registro
	username = r.FormValue("username")
	password := r.FormValue("password")

	// Crear un nuevo usuario
	newUser := models.User{
		Username: username,
		Password: utils.HashPassword(password),
	}

	// Guardar el usuario en la base de datos
	err = db.InsertUser(&newUser)
	if err != nil {
		log.Println("Error al insertar el usuario en la base de datos:", err)
		// Manejar error
		return
	}

	// Iniciar sesión automáticamente después del registro
	utils.StartSession(w, r, username)

	// Redirigir al dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
