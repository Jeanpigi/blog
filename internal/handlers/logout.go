package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Eliminar la sesión del usuario
	utils.EndSession(w, r)

	// Redirigir al inicio de sesión u otra página
	http.Redirect(w, r, "/login", http.StatusFound)
}
