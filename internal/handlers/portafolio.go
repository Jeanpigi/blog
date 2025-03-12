package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func PortafolioHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/portafolio.html", nil)
}
