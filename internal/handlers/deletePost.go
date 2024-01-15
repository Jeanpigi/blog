package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/Jeanpigi/blog/db"
	"github.com/gorilla/mux"
)

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	// Obtén el postID de la URL, por ejemplo, "/delete-post/123"
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		http.Error(w, "Invalid postID", http.StatusBadRequest)
		return
	}

	// Realiza la eliminación del post en la base de datos
	err = db.DeletePost(postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Envía una respuesta con un mensaje de éxito en formato JSON
	response := map[string]string{
		"message": "Post deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
