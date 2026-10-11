package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/utils"
)

// writeJSON serializa v como JSON con el status indicado.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError responde {"error": msg}. Todas las rutas /api/ usan este formato
// para que el JS del dashboard pueda mostrar el mensaje tal cual.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

const (
	maxTitleLen       = 255
	maxDescriptionLen = 500
	maxPostBody       = 1 << 20 // 1 MB (Quill incrusta imágenes pegadas en base64)
)

// postInput es el payload que aceptan crear/actualizar post. El autor nunca se
// lee del body: sale de la sesión.
type postInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Categoria   string `json:"categoria"`
}

// decodePostInput lee, normaliza y valida el body de un post.
func decodePostInput(w http.ResponseWriter, r *http.Request) (*models.Post, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPostBody)
	var in postInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, errors.New("el post supera 1 MB (¿imágenes muy pesadas pegadas en el editor?)")
		}
		return nil, errors.New("JSON inválido")
	}

	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Content = strings.TrimSpace(in.Content)

	switch {
	case in.Title == "":
		return nil, errors.New("el título es obligatorio")
	case utf8.RuneCountInString(in.Title) > maxTitleLen:
		return nil, errors.New("el título es demasiado largo")
	case in.Description == "":
		return nil, errors.New("la descripción es obligatoria")
	case utf8.RuneCountInString(in.Description) > maxDescriptionLen:
		return nil, errors.New("la descripción es demasiado larga")
	case isEmptyEditorHTML(in.Content):
		return nil, errors.New("el contenido es obligatorio")
	}

	cat, ok := normalizeCategory(in.Categoria)
	if !ok {
		return nil, errors.New("categoría inválida (usa Tech o Historias)")
	}

	return &models.Post{
		Title:       in.Title,
		Description: in.Description,
		// Se sanea también al guardar (además de al renderizar): el JSON de
		// /api/posts/{id} se carga en el editor del dashboard vía innerHTML.
		Content:   utils.SanitizePostHTML(template.HTML(in.Content)), //nolint:gosec // saneado por bluemonday
		Categoria: cat,
	}, nil
}

// normalizeCategory acepta mayúsculas/minúsculas ("tech" → "Tech").
func normalizeCategory(s string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "tech":
		return models.CategoryTech, true
	case "historias":
		return models.CategoryHistorias, true
	}
	return "", false
}

// isEmptyEditorHTML detecta el contenido "vacío" que deja Quill (<p><br></p>).
func isEmptyEditorHTML(s string) bool {
	r := strings.NewReplacer("<p>", "", "</p>", "", "<br>", "", "<br/>", "", "&nbsp;", "")
	return strings.TrimSpace(r.Replace(s)) == ""
}
