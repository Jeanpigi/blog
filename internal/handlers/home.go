package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "templates/home.html", nil)
}
