package models

import "html/template"

type Post struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Content     template.HTML `json:"content"`
	AuthorID    int           `json:"author_id"`
	CreatedAt   string        `json:"created_at"`
	Categoria   string        `json:"categoria"`
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
}
