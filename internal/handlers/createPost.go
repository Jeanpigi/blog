package handlers

import (
	"log"
	"net/http"
	"time"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/middleware"
)

// CreatePostHandler — POST /api/posts (requiere auth + CSRF).
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	post, translations, err := decodePostInput(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	post.AuthorID = user.ID
	post.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	id, err := db.InsertPost(post)
	if err != nil {
		log.Println("❌ Error creando post:", err)
		writeError(w, http.StatusInternalServerError, "No se pudo crear el post")
		return
	}
	post.ID = int(id)

	if err := saveTranslations(post.ID, translations); err != nil {
		log.Println("❌ Error guardando traducción:", err)
		writeError(w, http.StatusInternalServerError, "El post se creó, pero no se pudo guardar la versión en inglés. Edítalo para reintentar.")
		return
	}

	writeJSON(w, http.StatusCreated, post)
}
