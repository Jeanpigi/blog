package handlers

import (
	"net/http"
	"regexp"

	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

// Expresión regular para validar usernames
var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

// Mapa para rastrear intentos de login por IP
var loginAttempts = make(map[string]int)

// LoginHandler maneja el inicio de sesión
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr

	// Bloquear IP después de 5 intentos fallidos
	if loginAttempts[ip] >= 5 {
		http.Error(w, "Demasiados intentos. Intenta más tarde.", http.StatusTooManyRequests)
		return
	}

	sessionID := session.GetSessionID(r)

	if r.Method == "GET" {
		// Generar token CSRF
		csrfToken := utils.GenerateCSRFToken(sessionID)

		// Renderizar formulario con el token
		utils.RenderTemplate(w, "templates/login.html", map[string]interface{}{
			"CsrfToken": csrfToken,
		})
		return
	}

	// Validar token CSRF
	csrfToken := r.FormValue("csrf_token")
	if !utils.ValidateCSRF(sessionID, csrfToken) {
		http.Error(w, "CSRF token inválido", http.StatusForbidden)
		return
	}

	// Obtener datos del formulario
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validar el nombre de usuario
	if !validUsername.MatchString(username) {
		http.Error(w, "El nombre de usuario solo puede contener letras, números y guion bajo (3-20 caracteres)", http.StatusBadRequest)
		return
	}

	// Verificar usuario y contraseña usando `AuthenticateUser()`
	if !utils.AuthenticateUser(username, password) {
		loginAttempts[ip]++ // Incrementar intentos fallidos
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}

	// Resetear intentos fallidos si el login es exitoso
	delete(loginAttempts, ip)

	// Iniciar sesión
	session.StartSession(w, r, username)

	// Redirigir al dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}



