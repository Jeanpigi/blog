package handlers

import (
	"encoding/json"
	db "github.com/Jeanpigi/blog/db"
	"net/http"
)

func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Obtén todos los posts de la base de datos
	posts, err := db.GetAllPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Envía la lista de posts en la respuesta como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
