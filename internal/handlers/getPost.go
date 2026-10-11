package handlers

import (
	"net/http"

	db "github.com/Jeanpigi/blog/db"
	"github.com/gorilla/mux"
)

// GetPostsHandler — GET /api/posts/{id}
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	post, err := db.FindPostByID(mux.Vars(r)["id"])
	if err != nil || post == nil {
		writeError(w, http.StatusNotFound, "Post no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, post)
}
