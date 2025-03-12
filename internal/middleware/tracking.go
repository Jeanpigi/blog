package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
)

// Estructura para la respuesta de la API de geolocalización
type GeoInfo struct {
	Country   string  `json:"country"`
	Region    string  `json:"regionName"`
	City      string  `json:"city"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	ISP       string  `json:"isp"`
	Query     string  `json:"query"`
}

// Función para obtener la ubicación de la IP desde la API externa
func getGeoInfo(ip string) (GeoInfo, error) {
	var geoInfo GeoInfo
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	resp, err := http.Get(apiURL)
	if err != nil {
		return geoInfo, err
	}
	defer resp.Body.Close()

	// Decodificar respuesta de la API
	err = json.NewDecoder(resp.Body).Decode(&geoInfo)
	if err != nil {
		log.Println("❌ Error al decodificar JSON de geolocalización:", err)
		return geoInfo, err
	}

	// **Imprimir la respuesta de la API en la consola**
	log.Println("📌 Datos de geolocalización recibidos:", geoInfo)

	return geoInfo, nil
}

// Middleware para registrar visitas
func TrackVisitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		userAgent := r.UserAgent()
		page := r.URL.Path

		// Obtener datos de geolocalización
		geoInfo, err := getGeoInfo(ip)
		if err != nil {
			log.Println("⚠️ No se pudo obtener la geolocalización:", err)
		}

		// Crear objeto visita
		visit := &models.Visit{
			IP:        ip,
			UserAgent: userAgent,
			Page:      page,
			Country:   geoInfo.Country,
			Region:    geoInfo.Region,
			City:      geoInfo.City,
			Latitude:  geoInfo.Latitude,
			Longitude: geoInfo.Longitude,
			ISP:       geoInfo.ISP,
		}

		// Guardar en la base de datos
		if err := db.InsertVisit(visit); err != nil {
			log.Println("❌ Error registrando visita:", err)
		}

		next.ServeHTTP(w, r)
	})
}
