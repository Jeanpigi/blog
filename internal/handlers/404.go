package handlers

import (
	"net/http"
	"github.com/Jeanpigi/blog/internal/utils"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	utils.RenderTemplate(w, "templates/404.html", nil)
}
