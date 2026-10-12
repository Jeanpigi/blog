package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/internal/site"
	"github.com/Jeanpigi/blog/internal/utils"
)

// SupportHandler — GET /apoyar y /en/apoyar.
func SupportHandler(w http.ResponseWriter, r *http.Request) {
	links := site.Load()
	if !links.SupportEnabled() {
		NotFoundHandler(w, r)
		return
	}
	utils.RenderTemplate(w, r, "templates/apoyar.html", map[string]interface{}{"Links": links})
}

// PrivacyHandler — GET /privacidad y /en/privacidad.
func PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "templates/privacidad.html", map[string]interface{}{"Links": site.Load()})
}

// ServicesHandler — GET /servicios y /en/servicios.
func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "templates/servicios.html", map[string]interface{}{"Links": site.Load()})
}
