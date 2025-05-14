package middleware

import (
	"net/http"

	"github.com/Jeanpigi/blog/session"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := session.Store.Get(r, "session-name")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		_, ok := sess.Values["username"].(string)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	}
}
