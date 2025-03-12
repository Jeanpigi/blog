package handlers

import (
	"net/http"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener la sesión actual
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		// Manejar el error, posiblemente redirigiendo al login
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Comprobar si el nombre de usuario está en la sesión
	username, ok := session.Values["username"].(string)

	if !ok {
		// Si no está, redirigir al login
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		// Si el usuario no existe, redirigir al login
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	usuario := struct {
		Username string
		ID       int
	}{
		Username: username,
		ID:       user.ID,
	}

	utils.RenderTemplate(w, "templates/dashboard.html", usuario)
}
