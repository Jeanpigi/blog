package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/session"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Eliminar la sesión del usuario
	session.EndSession(w, r)

	// Redirigir al inicio de sesión
	http.Redirect(w, r, "/login", http.StatusFound)
}

