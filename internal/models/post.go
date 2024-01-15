package models

type Post struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	AuthorID    int    `json:"author_id"`
	CreatedAt   string `json:"created_at"`
	Categoria   string `json:"categoria"`
}
