package handlers

import (
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/utils"
	"github.com/Jeanpigi/blog/session"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Store.Get(r, "session-name")
	username := sess.Values["username"].(string)

	user, _ := db.GetUserByUsername(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := struct {
		Username string
		ID       int
	}{
		Username: username,
		ID:       user.ID,
	}

	utils.RenderTemplate(w, "dashboard.html", data)
}

