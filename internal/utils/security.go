package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"golang.org/x/crypto/bcrypt"
)

// CSRFCookieName es el nombre de la cookie que transporta el token CSRF.
// El patrón usado es "double-submit cookie": el mismo token viaja en la cookie
// y en el formulario (campo oculto) o en la cabecera X-CSRF-Token (fetch).
// El servidor solo confirma que ambos coinciden — no necesita estado en memoria.
const CSRFCookieName = "csrf_token"

// IssueCSRFToken genera un token aleatorio, lo deposita en una cookie y lo
// devuelve para incrustarlo en el formulario. La cookie NO es HttpOnly porque
// el JS del dashboard necesita leerla para reenviarla como cabecera en fetch().
func IssueCSRFToken(w http.ResponseWriter) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Println("Error al generar token CSRF:", err)
		return ""
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: false, // el JS lo lee para enviarlo en X-CSRF-Token
		SameSite: http.SameSiteStrictMode,
	})
	return token
}

// csrfCookieValue devuelve el valor de la cookie CSRF (o "" si no existe).
func csrfCookieValue(r *http.Request) string {
	c, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// constantTimeEqual compara sin filtrar timing.
func constantTimeEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateCSRFForm valida el token enviado por un formulario HTML (campo oculto
// "csrf_token") contra la cookie. Úsalo en POST de formularios (login, signup).
func ValidateCSRFForm(r *http.Request) bool {
	return constantTimeEqual(csrfCookieValue(r), r.FormValue("csrf_token"))
}

// ValidateCSRFHeader valida el token enviado por fetch en la cabecera
// X-CSRF-Token contra la cookie. NO toca el body, así que es seguro para
// endpoints JSON o multipart (el handler parsea el body después).
func ValidateCSRFHeader(r *http.Request) bool {
	return constantTimeEqual(csrfCookieValue(r), r.Header.Get("X-CSRF-Token"))
}

// 🔑 Hashea contraseñas de manera segura con bcrypt
func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal("Error al generar hash de contraseña:", err)
	}
	return string(hashedPassword)
}

// 🔓 Autenticación segura de usuarios
func AuthenticateUser(username, password string) bool {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		log.Println("❌ Error al obtener el usuario:", err)
		return false
	}
	if user == nil {
		log.Println("⚠️ Usuario no encontrado:", username)
		return false
	}

	// Comparar la contraseña hasheada almacenada con la ingresada
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("⚠️ Error al verificar la contraseña:", err)
		return false
	}

	log.Println("✅ Usuario autenticado correctamente:", username)
	return true
}




