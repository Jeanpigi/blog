package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Store.Get(r, "session-name")
	username, ok := sess.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Emite la cookie CSRF para que el JS del dashboard la reenvíe en X-CSRF-Token.
	csrfToken := utils.IssueCSRFToken(w)

	data := struct {
		Username  string
		ID        int
		CsrfToken string
	}{
		Username:  username,
		ID:        user.ID,
		CsrfToken: csrfToken,
	}

	utils.RenderTemplate(w, "templates/dashboard.html", data)
}

