package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/about.html", nil)
}
