package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
)

// GeoInfo representa la respuesta de ip-api.com
type GeoInfo struct {
	Country   string  `json:"country"`
	Region    string  `json:"regionName"`
	City      string  `json:"city"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	ISP       string  `json:"isp"`
}

type geoCacheEntry struct {
	info    GeoInfo
	expires time.Time
}

var (
	geoCache   sync.Map
	geoClient  = &http.Client{Timeout: 3 * time.Second}
)

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = strings.TrimSpace(ip[:idx])
		}
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	// Quitar puerto si viene incluido (IPv4:port)
	if idx := strings.LastIndex(ip, ":"); idx != -1 && strings.Count(ip, ":") == 1 {
		ip = ip[:idx]
	}
	return ip
}

func getGeoInfo(ip string) GeoInfo {
	// Revisar caché primero
	if v, ok := geoCache.Load(ip); ok {
		entry := v.(geoCacheEntry)
		if time.Now().Before(entry.expires) {
			return entry.info
		}
		geoCache.Delete(ip)
	}

	var info GeoInfo
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=country,regionName,city,lat,lon,isp", ip)
	resp, err := geoClient.Get(url)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Println("⚠️ Error decodificando geo:", err)
		return info
	}

	geoCache.Store(ip, geoCacheEntry{info: info, expires: time.Now().Add(24 * time.Hour)})
	return info
}

// TrackVisitMiddleware registra la visita de forma asíncrona (no bloquea la respuesta).
func TrackVisitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		userAgent := r.UserAgent()
		page := r.URL.Path

		// Servir primero, registrar después
		next.ServeHTTP(w, r)

		go func() {
			geo := getGeoInfo(ip)
			visit := &models.Visit{
				IP:        ip,
				UserAgent: userAgent,
				Page:      page,
				Country:   geo.Country,
				Region:    geo.Region,
				City:      geo.City,
				Latitude:  geo.Latitude,
				Longitude: geo.Longitude,
				ISP:       geo.ISP,
				Timestamp: time.Now(),
			}
			if err := db.InsertVisit(visit); err != nil {
				log.Println("❌ Error registrando visita:", err)
			}
		}()
	})
}
