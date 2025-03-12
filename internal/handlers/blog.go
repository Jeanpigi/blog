package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/blog.html", nil)
}
