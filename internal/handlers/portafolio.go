package handlers

import (
	"net/http"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/utils"
)

func PortafolioHandler(w http.ResponseWriter, r *http.Request) {
	experiences, err := db.GetAllExperiences()
	if err != nil {
		http.Error(w, "Error cargando experiencias", http.StatusInternalServerError)
		return
	}

	education, err := db.GetAllEducation()
	if err != nil {
		http.Error(w, "Error cargando educación", http.StatusInternalServerError)
		return
	}

	projects, err := db.GetAllProjects()
	if err != nil {
		http.Error(w, "Error cargando proyectos", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Experiences": experiences,
		"Education":   education,
		"Projects":    projects,
	}

	utils.RenderTemplate(w, "templates/portafolio.html", data)
}
