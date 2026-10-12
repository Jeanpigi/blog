package middleware

import "net/http"

// StaticCache añade Cache-Control a los archivos estáticos.
//   - Con ?v=... (CSS/JS versionados con staticVer): 1 año e immutable. Cada
//     despliegue cambia la versión, así que el navegador nunca usa uno viejo.
//   - Sin versión (imágenes, favicon): 7 días.
func StaticCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=604800")
		}
		next.ServeHTTP(w, r)
	})
}
