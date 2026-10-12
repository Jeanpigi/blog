package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "templates/blog.html", nil)
}
