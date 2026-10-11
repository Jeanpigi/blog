package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/home.html", nil)
}
