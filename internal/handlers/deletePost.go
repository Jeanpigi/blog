package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/session"
	"github.com/gorilla/mux"
)

// DeletePostHandler protege la eliminación de posts para que solo el autor pueda hacerlo
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Store.Get(r, "session-name")
	username := sess.Values["username"].(string)

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

	post, err := db.FindPostByID(vars["postID"])
	if err != nil || post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if post.AuthorID != user.ID {
		http.Error(w, "Forbidden: You can only delete your own posts", http.StatusForbidden)
		return
	}

	err = db.DeletePost(postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
}

