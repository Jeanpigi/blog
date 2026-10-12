// Package i18n resuelve el idioma de cada petición y traduce los textos de la
// interfaz. El español vive en las URLs de siempre (/blog, /post/5) y el
// inglés bajo el prefijo /en (/en/blog, /en/post/5): cada idioma tiene su URL
// propia e indexable, y las URLs existentes no cambian.
package i18n

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const (
	ES = "es"
	EN = "en"
)

type ctxKey int

const (
	langKey ctxKey = iota
	pathKey
)

var postPathRx = regexp.MustCompile(`^/post/[0-9]{1,9}$`)

var publicPages = map[string]bool{
	"/": true, "/blog": true, "/historias": true, "/tecnologias": true, "/portafolio": true,
	"/servicios": true, "/apoyar": true, "/privacidad": true,
}

// IsPublicPage indica si una ruta (sin prefijo de idioma) es una página
// pública traducible. El admin (/dashboard, /login…) solo existe en español.
func IsPublicPage(p string) bool {
	return publicPages[p] || postPathRx.MatchString(p)
}

// SplitLang separa el prefijo de idioma de una ruta: "/en/blog" → ("en", "/blog").
func SplitLang(p string) (lang, rest string) {
	if p == "/en" || strings.HasPrefix(p, "/en/") {
		rest = strings.TrimPrefix(p, "/en")
		if rest == "" {
			rest = "/"
		}
		return EN, rest
	}
	return ES, p
}

// Middleware detecta el idioma por el prefijo /en, lo quita de la ruta para
// que el router use las mismas rutas en ambos idiomas y guarda el idioma en el
// contexto. Un /en/... que no sea una página pública se deja tal cual (404).
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang, rest := SplitLang(r.URL.Path)
		if lang == EN && rest != "/" && strings.HasSuffix(rest, "/") {
			if trimmed := strings.TrimSuffix(rest, "/"); IsPublicPage(trimmed) {
				// Destino fijo bajo /en y validado contra la lista de páginas: no es un open redirect.
				http.Redirect(w, r, "/en"+trimmed, http.StatusMovedPermanently) //nolint:gosec // ver comentario
				return
			}
		}
		if lang == EN && !IsPublicPage(rest) {
			lang, rest = ES, r.URL.Path
		}
		ctx := context.WithValue(r.Context(), langKey, lang)
		ctx = context.WithValue(ctx, pathKey, rest)
		if rest != r.URL.Path {
			r2 := r.Clone(ctx)
			r2.URL.Path = rest
			r2.URL.RawPath = ""
			next.ServeHTTP(w, r2)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromRequest devuelve el idioma de la petición (por defecto español).
func FromRequest(r *http.Request) string {
	if l, ok := r.Context().Value(langKey).(string); ok {
		return l
	}
	return ES
}

// PagePath devuelve la ruta de la página sin prefijo de idioma.
func PagePath(r *http.Request) string {
	if p, ok := r.Context().Value(pathKey).(string); ok {
		return p
	}
	return r.URL.Path
}

// Localize antepone el prefijo del idioma a una ruta interna: ("en", "/blog") → "/en/blog".
func Localize(lang, path string) string {
	if lang != EN {
		return path
	}
	if path == "/" {
		return "/en"
	}
	return "/en" + path
}

// Other devuelve el otro idioma.
func Other(lang string) string {
	if lang == EN {
		return ES
	}
	return EN
}

// T traduce una clave. Si falta en inglés usa el español; si falta en ambos,
// devuelve la clave (así un texto sin traducir se ve enseguida).
func T(lang, key string, args ...interface{}) string {
	msg, ok := catalog[lang][key]
	if !ok {
		if msg, ok = catalog[ES][key]; !ok {
			return key
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
