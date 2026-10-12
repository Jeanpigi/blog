package models

import "html/template"

// Categorías válidas de un post (coinciden con la columna Posts.Category).
const (
	CategoryTech      = "Tech"
	CategoryHistorias = "Historias"
)

type Post struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Content     template.HTML `json:"content"`
	AuthorID    int           `json:"author_id"`
	CreatedAt   string        `json:"created_at"`
	Categoria   string        `json:"categoria"`
	// Traducciones por idioma ("en"); el post en sí está en español.
	Translations map[string]*PostTranslation `json:"translations,omitempty"`
}

// PostTranslation es la versión de un post en otro idioma.
type PostTranslation struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Content     template.HTML `json:"content"`
}

// PostListItem es el DTO liviano para listados (sin Content completo).
type PostListItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorID    int    `json:"author_id"`
	CreatedAt   string `json:"created_at"`
	Categoria   string `json:"categoria"`
	ReadingMin  int    `json:"reading_min"`
	Lang        string `json:"lang"`   // idioma en que vienen título y descripción
	HasEN       bool   `json:"has_en"` // tiene versión en inglés
}
