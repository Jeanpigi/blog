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
			post, err := decodePostInput(httptest.NewRecorder(), r)
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
	post, err := decodePostInput(httptest.NewRecorder(), r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(post.Content), "<script") {
		t.Errorf("el contenido no se saneó: %q", post.Content)
	}
}
