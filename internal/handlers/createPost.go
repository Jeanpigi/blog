package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/session"
)

// CreatePostHandler permite solo a usuarios autenticados crear posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Store.Get(r, "session-name")
	username := sess.Values["username"].(string)

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	post.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	post.AuthorID = user.ID

	err = db.InsertPost(&post)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Post created successfully")
}

