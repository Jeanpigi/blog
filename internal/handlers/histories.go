package handlers

import (
	"encoding/json"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
)

// Handler para obtener posts por categoría
func GetPostsByHistoryHandler(w http.ResponseWriter, r *http.Request) {
	categoria := "Historias"

	posts, err := db.FindPostsByCategory(categoria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
