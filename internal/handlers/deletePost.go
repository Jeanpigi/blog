package handlers

import (
	"log"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
)

// DeletePostHandler — DELETE /api/posts/{id} (solo el autor).
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	post, ok := ownedPost(w, r)
	if !ok {
		return
	}

	if err := db.DeletePost(post.ID); err != nil {
		log.Println("❌ Error eliminando post:", err)
		writeError(w, http.StatusInternalServerError, "No se pudo eliminar el post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
