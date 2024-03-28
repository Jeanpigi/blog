package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func HistoriasHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/historias.html", nil)
}
