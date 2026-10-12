package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP devuelve la IP real del cliente.
//
// Cadena en producción: Cloudflare → nginx → app. nginx tiene `real_ip_header
// CF-Connecting-IP` con `set_real_ip_from` limitado a los rangos de Cloudflare,
// así que $remote_addr ya es la IP verdadera, y la reenvía en X-Real-IP.
//
// Por eso se lee X-Real-IP y NO CF-Connecting-IP: esta última llega tal cual
// desde el cliente y, si alguien se conecta directo al servidor (saltándose
// Cloudflare), puede inventarla para evadir el límite de login o falsear visitas.
func ClientIP(r *http.Request) string {
	ip := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if ip == "" {
		// Sin nginx delante (desarrollo local): la IP de la conexión.
		ip = r.RemoteAddr
	}
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return ip
}
