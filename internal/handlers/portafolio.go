package handlers

import (
	"log"
	"net/http"
	"time"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/i18n"
	"github.com/Jeanpigi/blog/internal/utils"
)

func PortafolioHandler(w http.ResponseWriter, r *http.Request) {
	experiences, err := db.GetAllExperiences()
	if err != nil {
		log.Println("❌ portafolio experiencias:", err)
		http.Error(w, "Error cargando experiencias", http.StatusInternalServerError)
		return
	}

	education, err := db.GetAllEducation()
	if err != nil {
		log.Println("❌ portafolio educación:", err)
		http.Error(w, "Error cargando educación", http.StatusInternalServerError)
		return
	}

	projects, err := db.GetAllProjects()
	if err != nil {
		log.Println("❌ portafolio proyectos:", err)
		http.Error(w, "Error cargando proyectos", http.StatusInternalServerError)
		return
	}

	lang := i18n.FromRequest(r)
	for i := range experiences {
		experiences[i] = experiences[i].Localize(lang)
	}
	for i := range education {
		education[i] = education[i].Localize(lang)
	}
	for i := range projects {
		projects[i] = projects[i].Localize(lang)
	}

	// Años desde la primera experiencia (para el resumen del encabezado).
	years := 0
	for _, e := range experiences {
		if e.StartYear > 0 && (years == 0 || time.Now().Year()-e.StartYear > years) {
			years = time.Now().Year() - e.StartYear
		}
	}

	utils.RenderTemplate(w, r, "templates/portafolio.html", map[string]interface{}{
		"Experiences": experiences,
		"Education":   education,
		"Projects":    projects,
		"YearsExp":    years,
	})
}
