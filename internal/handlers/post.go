package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/i18n"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/gorilla/mux"
)

// PostHandler — GET /post/{id} y /en/post/{id}. En inglés usa la traducción
// si existe; si no, muestra el original en español con un aviso.
func PostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := db.FindPostByID(mux.Vars(r)["id"])
	if err != nil || post == nil {
		NotFoundHandler(w, r)
		return
	}

	lang := i18n.FromRequest(r)
	contentLang := i18n.ES
	if lang != i18n.ES {
		translations, err := db.GetPostTranslations(post.ID)
		if err != nil {
			log.Println("⚠️ traducciones del post:", err)
		}
		if t := translations[lang]; t != nil {
			post.Title, post.Description, post.Content = t.Title, t.Description, t.Content
			contentLang = lang
		}
	}

	// El contenido (HTML del editor Quill) se sanea con bluemonday antes de
	// mostrarse, aunque ya se saneó al guardar.
	utils.RenderTemplate(w, r, "templates/post.html", map[string]interface{}{
		"Post":         post,
		"ContentHTML":  utils.SanitizePostHTML(post.Content),
		"ContentLang":  contentLang,
		"Untranslated": contentLang != lang,
	})
}
