package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

// SignupHandler maneja el registro de nuevos usuarios
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// 🚀 Verificar si el usuario ya ha iniciado sesión
	if !session.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == "GET" {
		utils.RenderTemplate(w, "templates/signup.html", nil)
		return
	}

	// Si es un POST, procesar el formulario
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Verificar si el usuario ya existe
	existingUser, _ := db.GetUserByUsername(username)
	if existingUser != nil {
		http.Error(w, "El usuario ya existe", http.StatusConflict)
		return
	}

	// Crear un nuevo usuario con la contraseña hasheada
	newUser := models.User{
		Username: username,
		Password: utils.HashPassword(password),
	}

	// Guardar el usuario en la base de datos
	err := db.InsertUser(&newUser)
	if err != nil {
		log.Println("Error al insertar el usuario en la base de datos:", err)
		http.Error(w, "Error al registrar el usuario", http.StatusInternalServerError)
		return
	}

	// Iniciar sesión automáticamente después del registro
	session.StartSession(w, r, username)

	// Redirigir al dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}


