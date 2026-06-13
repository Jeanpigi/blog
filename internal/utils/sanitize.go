package utils

import (
	"html/template"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	postPolicy     *bluemonday.Policy
	postPolicyOnce sync.Once
)

// buildPostPolicy define qué HTML se permite en el contenido de un post.
// El editor Quill produce: p, br, strong/b, em/i, u, s, listas, enlaces,
// imágenes (a veces como data URI base64) y bloques de código (pre.ql-syntax).
func buildPostPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// Quill incrusta imágenes pegadas como data:image/...;base64
	p.AllowDataURIImages()
	// Conservar clases de Quill para bloques de código y alineación
	p.AllowAttrs("class").OnElements("pre", "code", "span", "p", "div")
	// Los enlaces se abren en pestaña nueva de forma segura
	p.RequireNoFollowOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}

// SanitizePostHTML limpia el HTML del contenido de un post y lo devuelve como
// template.HTML listo para incrustar sin re-escapar. Es seguro contra XSS
// almacenado aunque el contenido provenga de la base de datos.
// Nota: models.Post.Content ya es template.HTML (antes se renderizaba SIN sanear,
// lo que era un XSS almacenado); aquí se sanea siempre antes de mostrarlo.
func SanitizePostHTML(content template.HTML) template.HTML {
	postPolicyOnce.Do(func() { postPolicy = buildPostPolicy() })
	return template.HTML(postPolicy.Sanitize(string(content))) //nolint:gosec // ya saneado por bluemonday
}
