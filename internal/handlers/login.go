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
	// 🔒 Redirigir si ya está autenticado
	if session.IsAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	ip := r.RemoteAddr

	// ⛔ Bloquear IP tras múltiples intentos
	if loginAttempts[ip] >= 5 {
		http.Error(w, "Demasiados intentos. Intenta más tarde.", http.StatusTooManyRequests)
		return
	}

	sessionID := session.GetSessionID(r)

	// 📥 Mostrar formulario (GET)
	if r.Method == http.MethodGet {
		csrfToken := utils.GenerateCSRFToken(sessionID)
		utils.RenderTemplate(w, "templates/login.html", map[string]interface{}{
			"CsrfToken": csrfToken,
		})
		return
	}

	// 🔐 Validar CSRF
	csrfToken := r.FormValue("csrf_token")
	if !utils.ValidateCSRF(sessionID, csrfToken) {
		http.Error(w, "CSRF token inválido", http.StatusForbidden)
		return
	}

	// 📄 Leer datos del formulario
	username := r.FormValue("username")
	password := r.FormValue("password")

	// ✅ Validar username
	if !validUsername.MatchString(username) {
		http.Error(w, "El nombre de usuario solo puede contener letras, números y guion bajo (3-20 caracteres)", http.StatusBadRequest)
		return
	}

	// 🔍 Verificar credenciales
	if !utils.AuthenticateUser(username, password) {
		loginAttempts[ip]++ // ❌ Aumentar contador
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}

	// ✅ Autenticación exitosa
	delete(loginAttempts, ip)
	session.StartSession(w, r, username)

	// 🎯 Redirigir al dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}




