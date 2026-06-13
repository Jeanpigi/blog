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
		csrfToken := utils.IssueCSRFToken(w)
		utils.RenderTemplate(w, "templates/signup.html", map[string]interface{}{
			"CsrfToken": csrfToken,
		})
		return
	}

	// Validar CSRF (double-submit cookie)
	if !utils.ValidateCSRFForm(r) {
		http.Error(w, "CSRF token inválido", http.StatusForbidden)
		return
	}

	// Si es un POST, procesar el formulario
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validar formato de usuario (letras, números, guion bajo; 3-20 caracteres)
	if !validUsername.MatchString(username) {
		http.Error(w, "Usuario inválido: 3-20 caracteres alfanuméricos o guion bajo", http.StatusBadRequest)
		return
	}

	// Exigir una contraseña mínimamente fuerte
	if len(password) < 8 {
		http.Error(w, "La contraseña debe tener al menos 8 caracteres", http.StatusBadRequest)
		return
	}

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


