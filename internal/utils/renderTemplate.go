package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
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
	}
)

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
	buf.WriteTo(w)
}
