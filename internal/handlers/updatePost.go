package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/gorilla/mux"
)

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Obtén el postID de la URL, por ejemplo, "/update-post/123"
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		http.Error(w, "Invalid postID", http.StatusBadRequest)
		return
	}

	// Decodifica los datos del post actualizado recibidos en el cuerpo de la solicitud
	var post models.Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid post data", http.StatusBadRequest)
		return
	}

	// Realiza la actualización del post en la base de datos
	err = db.UpdatePost(postID, &post)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Envía una respuesta con un mensaje de éxito en formato JSON
	response := map[string]string{
		"message": "Post updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
