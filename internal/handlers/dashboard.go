package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/middleware"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/music"
	"github.com/Jeanpigi/blog/internal/utils"
)

// DashboardHandler — GET /dashboard (requiere auth).
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	// Emite la cookie CSRF: el JS la reenvía en X-CSRF-Token y el formulario
	// de logout la manda como campo oculto.
	csrfToken := utils.IssueCSRFToken(w)

	utils.RenderTemplate(w, "templates/dashboard.html", struct {
		Username  string
		CsrfToken string
	}{
		Username:  user.Username,
		CsrfToken: csrfToken,
	})
}

type dashboardStats struct {
	Posts struct {
		Total     int `json:"total"`
		Tech      int `json:"tech"`
		Historias int `json:"historias"`
	} `json:"posts"`
	Visits db.VisitStats `json:"visits"`
	Radio  struct {
		Songs       int    `json:"songs"`
		NowPlaying  string `json:"now_playing"`
		StartedAtMs int64  `json:"started_at"`
	} `json:"radio"`
}

// DashboardStatsHandler — GET /api/admin/stats (requiere auth).
// Un fallo parcial (p. ej. la tabla Visits) no tumba el resto de métricas.
func DashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	var st dashboardStats
	var err error

	if st.Posts.Total, err = db.CountPosts(); err != nil {
		log.Println("⚠️ stats posts:", err)
	}
	if st.Posts.Tech, err = db.CountPostsByCategory(models.CategoryTech); err != nil {
		log.Println("⚠️ stats tech:", err)
	}
	if st.Posts.Historias, err = db.CountPostsByCategory(models.CategoryHistorias); err != nil {
		log.Println("⚠️ stats historias:", err)
	}
	if st.Visits, err = db.GetVisitStats(); err != nil {
		log.Println("⚠️ stats visitas:", err)
	}

	song, startedAt := GetBroadcastState()
	st.Radio.Songs = len(music.Snapshot())
	if song != "" {
		st.Radio.NowPlaying = song
		st.Radio.StartedAtMs = startedAt.UnixMilli()
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, st)
}
