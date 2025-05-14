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
	sess, _ := session.Store.Get(r, "session-name")
	username := sess.Values["username"].(string)

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		http.Error(w, "Invalid postID", http.StatusBadRequest)
		return
	}

	post, err := db.FindPostByID(vars["postID"])
	if err != nil || post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if post.AuthorID != user.ID {
		http.Error(w, "Forbidden: You can only update your own posts", http.StatusForbidden)
		return
	}

	var updatedPost models.Post
	err = json.NewDecoder(r.Body).Decode(&updatedPost)
	if err != nil {
		http.Error(w, "Invalid post data", http.StatusBadRequest)
		return
	}

	err = db.UpdatePost(postID, &updatedPost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post updated successfully",
	})
}

