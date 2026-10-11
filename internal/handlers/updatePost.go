package handlers

import (
	"log"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/middleware"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/gorilla/mux"
)

// UpdatePostHandler — PUT/PATCH /api/posts/{id} (solo el autor).
func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	existing, ok := ownedPost(w, r)
	if !ok {
		return
	}

	post, err := decodePostInput(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.UpdatePost(existing.ID, post); err != nil {
		log.Println("❌ Error actualizando post:", err)
		writeError(w, http.StatusInternalServerError, "No se pudo actualizar el post")
		return
	}

	post.ID = existing.ID
	post.AuthorID = existing.AuthorID
	post.CreatedAt = existing.CreatedAt
	writeJSON(w, http.StatusOK, post)
}

// ownedPost carga el post {id} y verifica que pertenezca al usuario de la
// sesión. Si no, ya escribió la respuesta de error y devuelve ok=false.
func ownedPost(w http.ResponseWriter, r *http.Request) (*models.Post, bool) {
	post, err := db.FindPostByID(mux.Vars(r)["id"])
	if err != nil || post == nil {
		writeError(w, http.StatusNotFound, "Post no encontrado")
		return nil, false
	}
	if post.AuthorID != middleware.CurrentUser(r).ID {
		writeError(w, http.StatusForbidden, "Solo puedes modificar tus propios posts")
		return nil, false
	}
	return post, true
}
