package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/session"
	"github.com/gorilla/mux"
)

// UpdatePostHandler permite actualizar un post solo si el usuario es el autor
func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar si el usuario está autenticado
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Obtener el postID de la URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		http.Error(w, "Invalid postID", http.StatusBadRequest)
		return
	}

	// Buscar el post en la base de datos
	post, err := db.FindPostByID(vars["postID"])
	if err != nil || post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Verificar si el usuario autenticado es el autor del post
	if post.AuthorID != user.ID {
		http.Error(w, "Forbidden: You can only update your own posts", http.StatusForbidden)
		return
	}

	// Decodificar los datos del post actualizado
	var updatedPost models.Post
	err = json.NewDecoder(r.Body).Decode(&updatedPost)
	if err != nil {
		http.Error(w, "Invalid post data", http.StatusBadRequest)
		return
	}

	// Actualizar el post en la base de datos
	err = db.UpdatePost(postID, &updatedPost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Enviar una respuesta con un mensaje de éxito en formato JSON
	response := map[string]string{
		"message": "Post updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
