package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

func HistoriasHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/historias.html", nil)
}
