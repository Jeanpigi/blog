package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/gorilla/mux"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["id"]

	post, err := db.FindPostByID(postID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Renderizar la plantilla para los detalles del post con los datos obtenidos
	utils.RenderTemplate(w, "templates/post.html", map[string]interface{}{
		"Post": post,
	})
}
