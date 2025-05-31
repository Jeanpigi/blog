package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

// getClientIP obtiene la IP real del visitante incluso detrás de Nginx y Cloudflare
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	// Si incluye puerto (por ejemplo 127.0.0.1:54321), quitarlo
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

// getGeoInfo consulta la IP en ip-api.com
func getGeoInfo(ip string) (GeoInfo, error) {
	var geoInfo GeoInfo
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	resp, err := http.Get(apiURL)
	if err != nil {
		return geoInfo, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&geoInfo)
	if err != nil {
		log.Println("❌ Error al decodificar JSON de geolocalización:", err)
		return geoInfo, err
	}

	log.Println("📌 Datos de geolocalización recibidos:", geoInfo)
	return geoInfo, nil
}

// TrackVisitMiddleware registra cada visita
func TrackVisitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		userAgent := r.UserAgent()
		page := r.URL.Path

		// Obtener ubicación geográfica
		geoInfo, err := getGeoInfo(ip)
		if err != nil {
			log.Println("⚠️ No se pudo obtener la geolocalización:", err)
		}

		// Crear el modelo Visit
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
			Timestamp: time.Now(), // si agregaste este campo en la tabla
		}

		// Guardar en base de datos
		if err := db.InsertVisit(visit); err != nil {
			log.Println("❌ Error registrando visita:", err)
		}

		next.ServeHTTP(w, r)
	})
}
