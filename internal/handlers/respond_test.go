package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodePostInput(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
		wantCat string
	}{
		{"válido", `{"title":"Hola","description":"Desc","content":"<p>x</p>","categoria":"Tech"}`, "", "Tech"},
		{"categoría en minúsculas", `{"title":"Hola","description":"Desc","content":"<p>x</p>","categoria":"historias"}`, "", "Historias"},
		{"categoría inválida", `{"title":"Hola","description":"Desc","content":"<p>x</p>","categoria":"Otra"}`, "categoría", ""},
		{"sin título", `{"title":"  ","description":"Desc","content":"<p>x</p>","categoria":"Tech"}`, "título", ""},
		{"sin descripción", `{"title":"Hola","description":"","content":"<p>x</p>","categoria":"Tech"}`, "descripción", ""},
		{"editor vacío de Quill", `{"title":"Hola","description":"Desc","content":"<p><br></p>","categoria":"Tech"}`, "contenido", ""},
		{"JSON roto", `{`, "JSON", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/posts", strings.NewReader(tt.body))
			post, _, err := decodePostInput(httptest.NewRecorder(), r)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, se esperaba que contuviera %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if post.Categoria != tt.wantCat {
				t.Errorf("Categoria = %q, want %q", post.Categoria, tt.wantCat)
			}
		})
	}
}

func TestDecodePostInputSanitizesContent(t *testing.T) {
	body := `{"title":"T","description":"D","content":"<p>ok</p><script>alert(1)</script>","categoria":"Tech"}`
	r := httptest.NewRequest("POST", "/api/posts", strings.NewReader(body))
	post, _, err := decodePostInput(httptest.NewRecorder(), r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(post.Content), "<script") {
		t.Errorf("el contenido no se saneó: %q", post.Content)
	}
}

func TestDecodePostInputTranslations(t *testing.T) {
	base := `"title":"Hola","description":"Desc","content":"<p>x</p>","categoria":"Tech"`
	tests := []struct {
		name       string
		tr         string
		wantErr    bool
		wantDelete bool
		wantSave   bool
	}{
		{"sin traducción", ``, false, false, false},
		{"completa", `,"translations":{"en":{"title":"Hi","description":"D","content":"<p>y</p>"}}`, false, false, true},
		{"vacía = borrar", `,"translations":{"en":{"title":"","description":"","content":"<p><br></p>"}}`, false, true, false},
		{"a medias", `,"translations":{"en":{"title":"Hi","description":"","content":""}}`, true, false, false},
		{"idioma no soportado se ignora", `,"translations":{"fr":{"title":"Salut","description":"D","content":"<p>z</p>"}}`, false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/posts", strings.NewReader("{"+base+tt.tr+"}"))
			_, tr, err := decodePostInput(httptest.NewRecorder(), r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			en, present := tr["en"]
			if tt.wantDelete && (!present || en != nil) {
				t.Errorf("se esperaba borrar la traducción, got present=%v en=%v", present, en)
			}
			if tt.wantSave && (en == nil || en.Title != "Hi") {
				t.Errorf("se esperaba guardar la traducción, got %+v", en)
			}
			if !tt.wantDelete && !tt.wantSave && present {
				t.Errorf("no se esperaba tocar la traducción, got %+v", en)
			}
			if _, ok := tr["fr"]; ok {
				t.Error("un idioma no soportado no debe aceptarse")
			}
		})
	}
}
