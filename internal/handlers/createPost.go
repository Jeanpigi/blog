package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/session"
)

// CreatePostHandler permite solo a usuarios autenticados crear posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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

	// Decodifica los datos del post recibidos en el cuerpo de la solicitud
	var post models.Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Println("Error decoding post data:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Agrega la fecha y hora actual como CreatedAt
	post.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	// Asigna el ID del autor autenticado
	post.AuthorID = user.ID

	// Inserta el nuevo post en la base de datos utilizando la función InsertPost
	err = db.InsertPost(&post)
	if err != nil {
		log.Println("Error inserting post into database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Envía una respuesta de éxito al cliente
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Post created successfully")
}
