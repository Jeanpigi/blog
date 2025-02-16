package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jeanpigi/blog/db"
)

// Obtener visitas con ubicación
func GetVisitsWithLocationHandler(w http.ResponseWriter, r *http.Request) {
	visits, err := db.GetAllVisits()
	if err != nil {
		http.Error(w, "Error obteniendo visitas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visits)
}
