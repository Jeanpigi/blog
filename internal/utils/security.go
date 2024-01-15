package utils

import (
	"log"
	"net/http"
	"os"

	db "github.com/Jeanpigi/blog/db"
	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"
)

// Obtener la clave secreta desde la variable de entorno
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}
func AuthenticateUser(username, password string) bool {
	// Obtener el usuario de la base de datos
	user, err := db.GetUserByUsername(username)
	if err != nil {
		log.Println("Error al obtener el usuario:", err)
		return false
	}
	if user == nil {
		return false
	}

	// Verificar la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("Error al verificar la contraseña:", err)
		return false
	}

	return true
}

func StartSession(w http.ResponseWriter, r *http.Request, username string) {
	session, _ := store.Get(r, "session-name")
	session.Values["username"] = username
	session.Save(r, w)
}

func IsAuthenticated(r *http.Request) bool {
	session, _ := store.Get(r, "session-name")
	_, ok := session.Values["username"]
	return ok
}

func EndSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Options.MaxAge = -1 // Establecer la edad máxima de la sesión a un valor negativo para eliminarla
	session.Save(r, w)
}
