package handlers

import (
	"encoding/json"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
	"github.com/gorilla/mux"
)

func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["id"]
	// Obtén todos los posts de la base de datos
	post, err := db.FindPostByID(postID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Envía la lista de posts en la respuesta como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
