package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/utils"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		utils.RenderTemplate(w, "templates/signup.html")
	} else if r.Method == "POST" {
		// Procesar el formulario de registro
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Crear un nuevo usuario
		user := models.User{
			Username: username,
			Password: utils.HashPassword(password),
		}

		// Guardar el usuario en la base de datos (implementación propia)
		err := db.InsertUser(&user)
		if err != nil {
			log.Println("Error al insertar el usuario en la base de datos:", err)
			// Manejar el error, por ejemplo, mostrar un mensaje de error al usuario
			return
		}

		// Iniciar sesión y redirigir al dashboard
		utils.StartSession(w, r, username)
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}
