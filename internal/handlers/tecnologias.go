package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func TecnologiasHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/tecnologias.html")
}
