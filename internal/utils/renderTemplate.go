package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"
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

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateCacheMu.RLock()
	t, ok := templateCache[tmpl]
	templateCacheMu.RUnlock()

	if !ok {
		var err error
		t, err = template.New("").Funcs(templateFuncs).ParseFiles("templates/layout.html", tmpl)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
		templateCacheMu.Lock()
		templateCache[tmpl] = t
		templateCacheMu.Unlock()
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, _ = buf.WriteTo(w)
}
