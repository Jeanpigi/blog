package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

func TecnologiasHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/tecnologias.html", nil)
}
