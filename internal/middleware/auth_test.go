package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Jeanpigi/blog/session"
)

func TestMain(m *testing.M) {
	// Claves de prueba (base64 de 64 y 32 bytes).
	os.Setenv("SESSION_AUTH_KEY", strings.Repeat("A", 88))
	os.Setenv("SESSION_ENC_KEY", strings.Repeat("A", 43)+"=")
	session.InitStore()
	os.Exit(m.Run())
}

func TestRequireAuthWithoutSession(t *testing.T) {
	called := false
	h := RequireAuth(func(w http.ResponseWriter, r *http.Request) { called = true })

	t.Run("API responde 401 JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest("POST", "/api/posts", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
	})

	t.Run("página redirige a /login", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest("GET", "/dashboard", nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login" {
			t.Fatalf("status = %d, Location = %q", rec.Code, rec.Header().Get("Location"))
		}
	})

	if called {
		t.Error("el handler protegido se ejecutó sin sesión")
	}
}
