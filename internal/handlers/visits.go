package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
)

// GetVisitsWithLocationHandler — GET /api/visits/location (requiere auth).
func GetVisitsWithLocationHandler(w http.ResponseWriter, r *http.Request) {
	visits, err := db.GetAllVisits()
	if err != nil {
		log.Println("❌ Error obteniendo visitas:", err)
		writeError(w, http.StatusInternalServerError, "Error obteniendo visitas")
		return
	}
	if visits == nil {
		visits = []models.Visit{} // [] en vez de null
	}
	writeJSON(w, http.StatusOK, visits)
}
