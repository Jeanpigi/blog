package middleware

import (
	"net/http"
	"os"

	"github.com/Jeanpigi/blog/internal/utils"
)

// contentSecurityPolicy está afinada a los recursos externos que realmente usa
// el sitio: Google Analytics, SweetAlert2 (jsdelivr), Quill (quilljs), Google
// Fonts y Font Awesome (cdnjs). nginx NO setea CSP, así que este es el único
// lugar donde vive — NO agregarlo también en nginx (quedaría duplicado).
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://*.googletagmanager.com https://cdn.jsdelivr.net https://cdn.quilljs.com; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com https://cdn.quilljs.com; " +
	"font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com; " +
	"img-src 'self' data: https:; " + // https: permite imágenes embebidas en posts (pasivas, no ejecutan JS)
	"media-src 'self'; " +
	"connect-src 'self' https://*.google-analytics.com https://*.analytics.google.com https://*.googletagmanager.com; " +
	"object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'"

// SecurityHeaders añade el CSP a cada respuesta.
//
// En producción nginx YA inyecta X-Frame-Options, X-Content-Type-Options,
// Referrer-Policy, Permissions-Policy y HSTS (conf.d/security.conf), así que la
// app NO los repite (causaría cabeceras duplicadas/en conflicto). El único header
// de seguridad ausente en todo el stack es CSP, y es el que ponemos aquí.
//
// Si CSP_REPORT_ONLY=1, se emite como Content-Security-Policy-Report-Only: el
// navegador NO bloquea nada, solo reporta violaciones en consola. Úsalo para
// validar la política en producción antes de forzarla (luego quita la variable).
func SecurityHeaders(next http.Handler) http.Handler {
	headerName := "Content-Security-Policy"
	if os.Getenv("CSP_REPORT_ONLY") == "1" {
		headerName = "Content-Security-Policy-Report-Only"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		if h.Get("Content-Security-Policy") == "" && h.Get("Content-Security-Policy-Report-Only") == "" {
			h.Set(headerName, contentSecurityPolicy)
		}
		next.ServeHTTP(w, r)
	})
}

// RequireCSRF protege endpoints que mutan estado y se llaman vía fetch().
// Verifica la cabecera X-CSRF-Token contra la cookie (double-submit) sin tocar
// el body, por lo que es seguro encadenarlo antes de handlers JSON o multipart.
func RequireCSRF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !utils.ValidateCSRFHeader(r) {
			http.Error(w, "CSRF token inválido o ausente", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
