package session

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// 🛡️ Variable global para manejar sesiones
var Store *sessions.CookieStore

// 🔧 Inicializa la sesión con una clave segura
func InitStore() {
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("Error: SESSION_KEY no está configurado en el entorno")
	}

	Store = sessions.NewCookieStore([]byte(sessionKey))

	// 🔒 Configuración segura de sesiones
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,  // 7 días de duración
		HttpOnly: true,       // Evita acceso desde JavaScript (protección XSS)
		Secure:   true,       // Solo HTTPS
		SameSite: http.SameSiteStrictMode, // Evita ataques CSRF
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
