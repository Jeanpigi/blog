package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/analytics"
	"github.com/Jeanpigi/blog/internal/middleware"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/session"
)

// La medición se hace desde el navegador (static/js/track.js) y no desde el
// servidor: así los bots que no ejecutan JS no cuentan, la navegación PJAX se
// mide igual que una carga normal y se puede saber cuánto tiempo se leyó.

const maxReadSeconds = 4 * 60 * 60 // tope de 4h por vista (pestañas olvidadas)

var (
	viewIDRx     = regexp.MustCompile(`^[0-9a-f]{32}$`)
	trackLimiter = analytics.NewLimiter(120, 10*time.Minute)
)

type trackStart struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Referrer string `json:"referrer"`
}

type trackEnd struct {
	ID      string `json:"id"`
	Seconds int    `json:"seconds"`
}

// shouldTrack aplica los filtros comunes. Siempre se responde 204 para no
// revelar al cliente por qué se descartó un evento.
func shouldTrack(r *http.Request) (ip string, ok bool) {
	if analytics.IsBot(r.UserAgent()) {
		return "", false
	}
	// Las visitas del propio admin no cuentan.
	if session.IsAuthenticated(r) {
		return "", false
	}
	ip = middleware.ClientIP(r)
	return ip, trackLimiter.Allow(ip)
}

// TrackStartHandler — POST /api/track: registra una vista de página.
func TrackStartHandler(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusNoContent)

	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var in trackStart
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return
	}
	if !viewIDRx.MatchString(in.ID) || !analytics.TrackablePath(in.Path) {
		return
	}
	ip, ok := shouldTrack(r)
	if !ok {
		return
	}

	visit := &models.Visit{
		ViewID:    in.ID,
		IP:        analytics.VisitorID(ip), // nunca se guarda la IP real
		UserAgent: truncate(r.UserAgent(), 500),
		Page:      in.Path,
		Referrer:  analytics.ReferrerHost(in.Referrer, r.Host),
		Device:    analytics.Device(r.UserAgent()),
	}
	// Se inserta ya (sin ubicación) para que el evento de fin siempre encuentre
	// la fila; la geolocalización, que puede tardar segundos, se completa después.
	if err := db.InsertVisit(visit); err != nil {
		log.Println("❌ Error registrando visita:", err)
		return
	}
	go func() {
		geo := analytics.LookupGeo(ip) // la IP real solo se usa en memoria para geolocalizar
		if geo.Country == "" {
			return
		}
		if err := db.UpdateVisitGeo(in.ID, geo.Country, geo.Region, geo.City, geo.Latitude, geo.Longitude); err != nil {
			log.Println("⚠️ Error guardando ubicación:", err)
		}
	}()
}

// TrackEndHandler — POST /api/track/end: guarda cuánto tiempo estuvo visible la página.
func TrackEndHandler(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusNoContent)

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var in trackEnd
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return
	}
	if !viewIDRx.MatchString(in.ID) || in.Seconds < 1 {
		return
	}
	if _, ok := shouldTrack(r); !ok {
		return
	}
	if in.Seconds > maxReadSeconds {
		in.Seconds = maxReadSeconds
	}
	if err := db.UpdateVisitDuration(in.ID, in.Seconds); err != nil {
		log.Println("⚠️ Error guardando duración:", err)
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
