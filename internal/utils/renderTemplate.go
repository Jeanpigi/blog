package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Jeanpigi/blog/internal/i18n"
)

var (
	templateCache   = map[string]*template.Template{}
	templateCacheMu sync.RWMutex

	// staticVer se fija una vez al arrancar el servidor.
	// Cada despliegue genera un valor distinto → el navegador descarga los assets frescos.
	staticVer = fmt.Sprintf("%d", time.Now().Unix())

	templateFuncs = template.FuncMap{
		// staticVer devuelve la versión para cache-busting: ?v={{staticVer}}
		"staticVer": func() string { return staticVer },
		"srcset":    imageSrcset,
		"hostOf":    hostOf,
		"initials":  initials,
	}
)

func init() {
	// Marcadores con la firma correcta; RenderTemplate los reemplaza por petición.
	for k, v := range requestFuncs(i18n.ES, "/") {
		templateFuncs[k] = v
	}
}

// imageSrcset genera el srcset de una imagen local optimizada. Las imágenes de
// static/img/ se guardan en dos anchos (nombre-480.webp y nombre-960.webp); si
// la URL apunta a la de 960 devuelve ambas para que el navegador elija. Para
// cualquier otra URL (externa, sin variantes) devuelve "".
func imageSrcset(src string) string {
	if !strings.HasPrefix(src, "/static/img/") || !strings.HasSuffix(src, "-960.webp") {
		return ""
	}
	small := strings.TrimSuffix(src, "-960.webp") + "-480.webp"
	return small + " 480w, " + src + " 960w"
}

// hostOf devuelve el dominio de una URL sin "www." (para mostrar "vercel.app").
func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}

// initials devuelve hasta dos iniciales de un título (placeholder de proyectos sin imagen).
func initials(s string) string {
	out := ""
	for _, w := range strings.Fields(s) {
		r := []rune(w)
		if len(r) > 0 && unicode.IsLetter(r[0]) {
			out += strings.ToUpper(string(r[0]))
		}
		if len([]rune(out)) == 2 {
			break
		}
	}
	return out
}

// SetStaticVersion permite sobreescribir la versión desde main.go
// (útil para fijar el hash de git en producción).
func SetStaticVersion(v string) {
	if v != "" {
		staticVer = v
	}
}

// siteURL es la URL pública del sitio, para enlaces absolutos (hreflang,
// Open Graph). Se puede cambiar con SITE_URL.
func siteURL() string {
	if u := os.Getenv("SITE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "https://blog.jbearp.com"
}

// RenderTemplate renderiza tmpl dentro del layout. Las plantillas se parsean
// una vez y se cachean SIN ejecutar; en cada petición se clona la base y se le
// enlazan las funciones de idioma (t, url, lang…) de esa petición.
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	templateCacheMu.RLock()
	base, ok := templateCache[tmpl]
	templateCacheMu.RUnlock()

	if !ok {
		var err error
		base, err = template.New("").Funcs(templateFuncs).ParseFiles("templates/layout.html", tmpl)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
		templateCacheMu.Lock()
		templateCache[tmpl] = base
		templateCacheMu.Unlock()
	}

	t, err := base.Clone()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	lang := i18n.FromRequest(r)
	page := i18n.PagePath(r)
	t.Funcs(requestFuncs(lang, page))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Language", lang)
	_, _ = buf.WriteTo(w)
}

// requestFuncs son las funciones de plantilla que dependen del idioma y de la
// página actual. En templateFuncs hay versiones vacías con los mismos nombres
// para que las plantillas se puedan parsear.
func requestFuncs(lang, page string) template.FuncMap {
	return template.FuncMap{
		"lang":     func() string { return lang },
		"t":        func(key string, args ...interface{}) string { return i18n.T(lang, key, args...) },
		"url":      func(p string) string { return i18n.Localize(lang, p) },
		"isPublic": func() bool { return i18n.IsPublicPage(page) },
		"langURL":  func(l string) string { return i18n.Localize(l, page) },
		"absURL":   func(p string) string { return siteURL() + p },
	}
}
