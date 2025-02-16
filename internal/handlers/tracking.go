package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

// Handler para mostrar la página de visitas con el mapa
func VisitsPageHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/tracking.html", nil)
}
