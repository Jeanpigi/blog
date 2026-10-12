package middleware

import (
	"net/http"
	"strings"
)

// ClientIP devuelve la IP real del cliente. Detrás de Cloudflare + nginx la IP
// llega en CF-Connecting-IP; r.RemoteAddr sería la del proxy (127.0.0.1).
// IMPORTANTE: confía en estas cabeceras solo porque el origen está restringido
// a las IPs de Cloudflare. Si el origen fuera accesible directamente, un atacante
// podría falsificarlas (p. ej. para evadir el rate-limit de login).
func ClientIP(r *http.Request) string {
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
