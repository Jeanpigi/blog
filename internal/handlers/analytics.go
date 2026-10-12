package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/Jeanpigi/blog/db"
)

var allowedPeriods = map[int]bool{7: true, 30: true, 90: true, 365: true}

type analyticsResponse struct {
	Days          int                 `json:"days"`
	MeasuredSince string              `json:"measured_since"`
	Summary       db.AnalyticsSummary `json:"summary"`
	Daily         []db.DailyPoint     `json:"daily"`
	TopPosts      []db.TopPost        `json:"top_posts"`
	Pages         []db.LabelCount     `json:"pages"`
	Countries     []db.LabelCount     `json:"countries"`
	Referrers     []db.LabelCount     `json:"referrers"`
	Devices       []db.LabelCount     `json:"devices"`
}

// AnalyticsHandler — GET /api/admin/analytics?days=30 (requiere auth).
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	days := parseInt(r.URL.Query().Get("days"), 30)
	if !allowedPeriods[days] {
		days = 30
	}

	res := analyticsResponse{Days: days}
	var err error
	fail := func(what string, err error) bool {
		if err != nil {
			log.Printf("❌ analytics %s: %v", what, err)
			writeError(w, http.StatusInternalServerError, "No se pudieron calcular las estadísticas")
			return true
		}
		return false
	}

	if res.Summary, err = db.GetAnalyticsSummary(days); fail("summary", err) {
		return
	}
	daily, err := db.GetDailyViews(days)
	if fail("daily", err) {
		return
	}
	today, err := db.Today()
	if fail("today", err) {
		return
	}
	res.Daily = fillDays(daily, today, days)
	if res.TopPosts, err = db.GetTopPosts(days, 10); fail("top posts", err) {
		return
	}
	for dim, dst := range map[string]*[]db.LabelCount{
		"page": &res.Pages, "country": &res.Countries, "referrer": &res.Referrers, "device": &res.Devices,
	} {
		if *dst, err = db.GetBreakdown(dim, days, 8); fail(dim, err) {
			return
		}
	}
	if res.MeasuredSince, err = db.MeasuredSince(); fail("measured since", err) {
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, res)
}

// fillDays completa con ceros los días sin visitas para que la gráfica tenga
// un punto por día (un hueco no es lo mismo que "no hay datos").
func fillDays(points []db.DailyPoint, today string, days int) []db.DailyPoint {
	end, err := time.Parse("2006-01-02", today)
	if err != nil {
		return points
	}
	byDate := make(map[string]db.DailyPoint, len(points))
	for _, p := range points {
		byDate[p.Date] = p
	}
	out := make([]db.DailyPoint, 0, days)
	for d := end.AddDate(0, 0, -(days - 1)); !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		p, ok := byDate[key]
		if !ok {
			p = db.DailyPoint{Date: key}
		}
		out = append(out, p)
	}
	return out
}
