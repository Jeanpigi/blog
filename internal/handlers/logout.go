package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

// LogoutHandler — POST /logout. Solo POST + CSRF: con GET cualquier sitio
// podría cerrar la sesión del admin con un simple <img src="/logout">.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if !utils.ValidateCSRFForm(r) {
		http.Error(w, "CSRF token inválido. Recarga la página e intenta de nuevo.", http.StatusForbidden)
		return
	}
	session.EndSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
