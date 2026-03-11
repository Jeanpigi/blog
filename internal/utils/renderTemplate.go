package utils

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"sync"
)

var (
	templateCache   = map[string]*template.Template{}
	templateCacheMu sync.RWMutex
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateCacheMu.RLock()
	t, ok := templateCache[tmpl]
	templateCacheMu.RUnlock()

	if !ok {
		var err error
		t, err = template.ParseFiles("templates/layout.html", tmpl)
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
