package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	var gotPath, gotLang, gotPage string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotLang, gotPage = r.URL.Path, FromRequest(r), PagePath(r)
	}))
	cases := []struct{ in, path, lang string }{
		{"/", "/", ES},
		{"/blog", "/blog", ES},
		{"/en", "/", EN},
		{"/en/blog", "/blog", EN},
		{"/en/post/12", "/post/12", EN},
		{"/en/dashboard", "/en/dashboard", ES}, // el admin no existe en inglés → 404
		{"/english", "/english", ES},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", c.in, nil))
		if gotPath != c.path || gotLang != c.lang {
			t.Errorf("%s → path=%q lang=%q, want %q %q", c.in, gotPath, gotLang, c.path, c.lang)
		}
		if c.lang == EN && gotPage != c.path {
			t.Errorf("%s → PagePath=%q", c.in, gotPage)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/en/blog/", nil))
	if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/en/blog" {
		t.Errorf("/en/blog/ → %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestLocalizeAndT(t *testing.T) {
	if Localize(EN, "/") != "/en" || Localize(EN, "/blog") != "/en/blog" || Localize(ES, "/blog") != "/blog" {
		t.Error("Localize incorrecto")
	}
	if T(EN, "nav.home") != "Home" || T(ES, "nav.home") != "Inicio" {
		t.Error("T no traduce")
	}
	if T(EN, "clave.inexistente") != "clave.inexistente" {
		t.Error("una clave faltante debe devolverse tal cual")
	}
	if T(EN, "post.shareOn", "X") != "Share on X" {
		t.Errorf("T con argumentos: %q", T(EN, "post.shareOn", "X"))
	}
}

// Cada clave en español debe existir en inglés (y viceversa).
func TestCatalogComplete(t *testing.T) {
	for k := range catalog[ES] {
		if _, ok := catalog[EN][k]; !ok {
			t.Errorf("falta en inglés: %s", k)
		}
	}
	for k := range catalog[EN] {
		if _, ok := catalog[ES][k]; !ok {
			t.Errorf("falta en español: %s", k)
		}
	}
}
