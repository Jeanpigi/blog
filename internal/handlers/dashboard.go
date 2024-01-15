package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar si el usuario está autenticado
	if !utils.IsAuthenticated(r) {
		// Redirigir al inicio de sesión si no está autenticado
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Tu código para manejar el dashboard
	utils.RenderTemplate(w, "templates/dashboard.html")
}
