// Package analytics contiene la lógica de medición de visitas: qué se cuenta
// (páginas públicas, sin bots), cómo se clasifica (dispositivo, referrer) y la
// geolocalización por IP.
package analytics

import (
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// botPattern detecta rastreadores y clientes no humanos por su User-Agent.
// Se usa también en SQL (BotSQLPattern) para limpiar los datos históricos,
// que se registraban desde el servidor e incluían bots.
const BotSQLPattern = `bot|crawl|spider|slurp|scrap|fetch|monitor|uptime|preview|headless|lighthouse|pingdom|curl|wget|python|go-http|java/|okhttp|axios|node-fetch|libwww|httpclient|facebookexternalhit|whatsapp|telegram|discord|semrush|ahrefs|mj12|bytespider|gptbot|claudebot`

var botRx = regexp.MustCompile(`(?i)` + BotSQLPattern)

// IsBot devuelve true si el User-Agent parece de un bot (o está vacío).
func IsBot(ua string) bool {
	return strings.TrimSpace(ua) == "" || botRx.MatchString(ua)
}

// Device clasifica el User-Agent en "mobile", "tablet" o "desktop".
func Device(ua string) string {
	u := strings.ToLower(ua)
	switch {
	case strings.Contains(u, "ipad") || strings.Contains(u, "tablet") ||
		(strings.Contains(u, "android") && !strings.Contains(u, "mobile")):
		return "tablet"
	case strings.Contains(u, "mobi") || strings.Contains(u, "iphone") || strings.Contains(u, "android"):
		return "mobile"
	default:
		return "desktop"
	}
}

// ReferrerHost devuelve el dominio externo desde el que llegó la visita
// ("google.com", "instagram.com"...). Devuelve "" si es acceso directo o
// navegación interna dentro del propio sitio (ownHost).
func ReferrerHost(ref, ownHost string) string {
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil || u.Hostname() == "" {
		return ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if host == strings.TrimPrefix(strings.ToLower(stripPort(ownHost)), "www.") {
		return ""
	}
	// Agrupa variantes conocidas (l.instagram.com, m.facebook.com, google.com.co…).
	for _, base := range []string{"google", "bing", "duckduckgo", "facebook", "instagram", "youtube", "linkedin", "reddit", "github"} {
		if host == base+".com" || strings.HasSuffix(host, "."+base+".com") || strings.HasPrefix(host, base+".") {
			return base + ".com"
		}
	}
	switch host {
	case "t.co", "x.com", "twitter.com":
		return "x.com"
	}
	if len(host) > 100 {
		host = host[:100]
	}
	return host
}

func stripPort(h string) string {
	if i := strings.LastIndex(h, ":"); i != -1 && !strings.Contains(h[i:], "]") {
		return h[:i]
	}
	return h
}

var (
	postPathRx  = regexp.MustCompile(`^/post/[0-9]{1,9}$`)
	publicPaths = map[string]bool{
		"/": true, "/blog": true, "/historias": true, "/tecnologias": true, "/portafolio": true,
	}
)

// TrackablePath indica si una ruta es una página pública que se mide.
// Las páginas de admin (/dashboard, /visitas, /login...) nunca se cuentan.
func TrackablePath(p string) bool {
	return publicPaths[p] || postPathRx.MatchString(p)
}

// ── Límite de eventos por IP ───────────────────────────────────────────────

// Limiter es un límite simple de ventana fija por clave (IP). Evita que alguien
// infle las estadísticas enviando eventos falsos en bucle.
type Limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string]*bucket
	sweep  time.Time
}

type bucket struct {
	count int
	reset time.Time
}

// NewLimiter permite max eventos por clave cada window.
func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, hits: map[string]*bucket{}}
}

// Allow registra un evento para key y devuelve false si superó el límite.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	// Limpieza ocasional para que el mapa no crezca sin control.
	if now.After(l.sweep) {
		for k, b := range l.hits {
			if now.After(b.reset) {
				delete(l.hits, k)
			}
		}
		l.sweep = now.Add(l.window)
	}

	b, ok := l.hits[key]
	if !ok || now.After(b.reset) {
		l.hits[key] = &bucket{count: 1, reset: now.Add(l.window)}
		return true
	}
	b.count++
	return b.count <= l.max
}
