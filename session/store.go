package session

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// Variable global para manejar sesiones
var Store *sessions.CookieStore

// InitStore inicializa el store con dos claves: autenticación (HMAC) y cifrado (AES).
// SESSION_AUTH_KEY: 64 bytes en base64 (para HMAC-SHA512)
// SESSION_ENC_KEY:  32 bytes en base64 (para AES-256)
func InitStore() {
	authKeyB64 := os.Getenv("SESSION_AUTH_KEY")
	encKeyB64 := os.Getenv("SESSION_ENC_KEY")

	if authKeyB64 == "" || encKeyB64 == "" {
		log.Fatal("Error: SESSION_AUTH_KEY y SESSION_ENC_KEY deben estar configurados")
	}

	authKey, err := base64.StdEncoding.DecodeString(authKeyB64)
	if err != nil || len(authKey) < 32 {
		log.Fatalf("SESSION_AUTH_KEY inválida: debe ser base64 de al menos 32 bytes (recomendado 64)")
	}

	encKey, err := base64.StdEncoding.DecodeString(encKeyB64)
	if err != nil || (len(encKey) != 16 && len(encKey) != 24 && len(encKey) != 32) {
		log.Fatalf("SESSION_ENC_KEY inválida: debe ser base64 de 16, 24 o 32 bytes")
	}

	// Dos claves: la primera firma (HMAC), la segunda cifra (AES)
	Store = sessions.NewCookieStore(authKey, encKey)

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// 🏷️ Iniciar una nueva sesión para el usuario
func StartSession(w http.ResponseWriter, r *http.Request, username string) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error al obtener la sesión:", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	session.Values["username"] = username
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error al guardar la sesión:", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// ✅ Verifica si un usuario está autenticado
func IsAuthenticated(r *http.Request) bool {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error al obtener la sesión:", err)
		return false
	}

	_, ok := session.Values["username"]
	return ok
}

// 🚪 Cierra la sesión del usuario
func EndSession(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error al obtener la sesión para cerrar:", err)
		return
	}

	session.Options.MaxAge = -1 // Eliminar la sesión
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error al cerrar la sesión:", err)
	}
}

// 🔍 Obtener el ID de sesión actual
func GetSessionID(r *http.Request) string {
	session, _ := Store.Get(r, "session-name")
	sessionID, ok := session.Values["sessionID"].(string)
	if !ok {
		return ""
	}
	return sessionID
}
