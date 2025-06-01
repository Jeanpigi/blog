// models/portfolio.go
package models

type Experience struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Place       string `json:"place"`
	Description string `json:"description"`
	StartYear   int    `json:"start_year"`
	EndYear     int    `json:"end_year"`
}

type Education struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Place     string `json:"place"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
}

type Project struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ProjectURL  string `json:"project_url"`
}
