package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/session"
)

type ctxKey int

const userKey ctxKey = iota

// RequireAuth exige una sesión válida y carga el usuario en el contexto del
// request (léelo con CurrentUser). Para rutas /api/ responde 401 JSON en vez
// de redirigir: un fetch() seguiría el 302 y recibiría el HTML del login con
// status 200, haciendo creer al JS que la operación tuvo éxito.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := sessionUser(r)
		if user == nil {
			unauthorized(w, r)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	}
}

// CurrentUser devuelve el usuario autenticado. Solo es válido dentro de
// handlers envueltos por RequireAuth (en otro caso devuelve nil).
func CurrentUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(userKey).(*models.User)
	return u
}

func sessionUser(r *http.Request) *models.User {
	sess, err := session.Store.Get(r, session.Name)
	if err != nil {
		return nil
	}
	username, ok := sess.Values["username"].(string)
	if !ok || username == "" {
		return nil
	}
	// Se consulta la DB para que borrar un usuario invalide sus sesiones activas.
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return nil
	}
	return user
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "No autenticado"})
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
