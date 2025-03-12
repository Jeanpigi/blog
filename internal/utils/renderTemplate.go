package utils

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		"templates/layout.html",
		tmpl,
	}

	templates, err := template.ParseFiles(files...)

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var buf bytes.Buffer
	err = templates.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	buf.WriteTo(w) // Solo se escribe en w si todo salió bien

}
