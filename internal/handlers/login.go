package handlers

import (
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

// Expresión regular para validar usernames
var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

const (
	maxLoginAttempts = 5
	loginBanDuration = 15 * time.Minute
)

type loginEntry struct {
	attempts int
	expiry   time.Time
}

var (
	loginMu      sync.Mutex
	loginAttempts = make(map[string]loginEntry)
)

// remoteIP extrae solo la IP de r.RemoteAddr (que viene como "ip:port").
func remoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isBlocked devuelve true si la IP sigue bloqueada.
func isBlocked(ip string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()
	entry, ok := loginAttempts[ip]
	if !ok {
		return false
	}
	if time.Now().After(entry.expiry) {
		delete(loginAttempts, ip)
		return false
	}
	return entry.attempts >= maxLoginAttempts
}

// recordFailure incrementa el contador; reinicia el TTL en el primer fallo.
func recordFailure(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	entry := loginAttempts[ip]
	entry.attempts++
	if entry.attempts == 1 {
		entry.expiry = time.Now().Add(loginBanDuration)
	}
	loginAttempts[ip] = entry
}

// clearAttempts elimina el registro de una IP tras login exitoso.
func clearAttempts(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginAttempts, ip)
}

// LoginHandler maneja el inicio de sesión.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Redirigir si ya está autenticado
	if session.IsAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	ip := remoteIP(r)

	// Mostrar formulario (GET)
	if r.Method == http.MethodGet {
		sessionID := session.GetSessionID(r)
		csrfToken := utils.GenerateCSRFToken(sessionID)
		utils.RenderTemplate(w, "templates/login.html", map[string]interface{}{
			"CsrfToken": csrfToken,
		})
		return
	}

	// Bloquear IP tras múltiples intentos
	if isBlocked(ip) {
		http.Redirect(w, r, "/login?error=too_many_attempts", http.StatusSeeOther)
		return
	}

	// Validar CSRF
	sessionID := session.GetSessionID(r)
	csrfToken := r.FormValue("csrf_token")
	if !utils.ValidateCSRF(sessionID, csrfToken) {
		http.Redirect(w, r, "/login?error=csrf_invalid", http.StatusSeeOther)
		return
	}

	// Leer datos del formulario
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validar formato de username
	if !validUsername.MatchString(username) {
		http.Redirect(w, r, "/login?error=invalid_username", http.StatusSeeOther)
		return
	}

	// Verificar credenciales
	if !utils.AuthenticateUser(username, password) {
		recordFailure(ip)
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}

	// Autenticación exitosa
	clearAttempts(ip)
	session.StartSession(w, r, username)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
