package db

import (
	"database/sql"
	"fmt"

	"github.com/Jeanpigi/blog/internal/analytics"
)

// Todas las consultas de estadísticas excluyen bots por User-Agent. Los datos
// históricos se registraban desde el servidor e incluían rastreadores; con este
// filtro también quedan limpios.
const humanVisits = `timestamp >= NOW() - INTERVAL ? DAY AND NOT REGEXP_LIKE(user_agent, ?, 'i')`

// measured limita a las vistas registradas con la medición client-side (tienen
// view_id): son las únicas con referrer, dispositivo y duración fiables.
const measured = ` AND view_id IS NOT NULL`

// UpdateVisitDuration guarda el tiempo de lectura de una vista. Solo crece
// (GREATEST) y solo para vistas recientes, así un id filtrado no sirve para
// alterar datos antiguos.
func UpdateVisitDuration(viewID string, seconds int) error {
	const q = `UPDATE Visits SET duration_sec = GREATEST(duration_sec, ?)
	           WHERE view_id = ? AND timestamp >= NOW() - INTERVAL 6 HOUR`
	_, err := Db.Exec(q, seconds, viewID)
	return err
}

type AnalyticsSummary struct {
	Views          int     `json:"views"`
	Visitors       int     `json:"visitors"`
	AvgDurationSec float64 `json:"avg_duration_sec"`
	PrevViews      int     `json:"prev_views"`
	PrevVisitors   int     `json:"prev_visitors"`
}

type DailyPoint struct {
	Date     string `json:"date"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type TopPost struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	Views          int     `json:"views"`
	Visitors       int     `json:"visitors"`
	AvgDurationSec float64 `json:"avg_duration_sec"`
}

type LabelCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetAnalyticsSummary devuelve totales del periodo y del periodo anterior
// (mismo largo) para mostrar la variación.
func GetAnalyticsSummary(days int) (AnalyticsSummary, error) {
	var s AnalyticsSummary
	err := Db.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT ip), COALESCE(AVG(NULLIF(duration_sec, 0)), 0)
		FROM Visits WHERE `+humanVisits, days, analytics.BotSQLPattern,
	).Scan(&s.Views, &s.Visitors, &s.AvgDurationSec)
	if err != nil {
		return s, err
	}
	err = Db.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT ip) FROM Visits
		WHERE timestamp >= NOW() - INTERVAL ? DAY AND timestamp < NOW() - INTERVAL ? DAY
		  AND NOT REGEXP_LIKE(user_agent, ?, 'i')`, days*2, days, analytics.BotSQLPattern,
	).Scan(&s.PrevViews, &s.PrevVisitors)
	return s, err
}

// GetDailyViews devuelve vistas y visitantes por día (solo días con datos;
// el handler rellena los huecos).
func GetDailyViews(days int) ([]DailyPoint, error) {
	rows, err := Db.Query(`
		SELECT DATE_FORMAT(timestamp, '%Y-%m-%d') d, COUNT(*), COUNT(DISTINCT ip)
		FROM Visits WHERE `+humanVisits+`
		GROUP BY d ORDER BY d`, days, analytics.BotSQLPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DailyPoint
	for rows.Next() {
		var p DailyPoint
		if err := rows.Scan(&p.Date, &p.Views, &p.Visitors); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetTopPosts devuelve los posts más vistos con su tiempo medio de lectura.
func GetTopPosts(days, limit int) ([]TopPost, error) {
	rows, err := Db.Query(`
		SELECT p.ID, p.Title, COUNT(*) views, COUNT(DISTINCT v.ip),
		       COALESCE(AVG(NULLIF(v.duration_sec, 0)), 0)
		FROM Visits v JOIN Posts p ON v.page IN (CONCAT('/post/', p.ID), CONCAT('/en/post/', p.ID))
		WHERE v.`+humanVisits+`
		GROUP BY p.ID, p.Title ORDER BY views DESC LIMIT ?`, days, analytics.BotSQLPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TopPost
	for rows.Next() {
		var t TopPost
		if err := rows.Scan(&t.ID, &t.Title, &t.Views, &t.Visitors, &t.AvgDurationSec); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Dimensiones permitidas para GetBreakdown (nunca se interpola input del usuario).
var breakdownColumns = map[string]struct {
	expr     string
	measured bool
}{
	"page":     {"page", false},
	"country":  {"COALESCE(NULLIF(country, ''), 'Desconocido')", false},
	"referrer": {"COALESCE(NULLIF(referrer, ''), 'Directo')", true},
	"device":   {"COALESCE(NULLIF(device, ''), 'Desconocido')", true},
}

// GetBreakdown cuenta vistas agrupadas por una dimensión (page, country,
// referrer, device). Las páginas excluyen /post/{id}: esas van en GetTopPosts.
func GetBreakdown(dim string, days, limit int) ([]LabelCount, error) {
	col, ok := breakdownColumns[dim]
	if !ok {
		return nil, fmt.Errorf("dimensión inválida: %s", dim)
	}
	where := humanVisits
	if col.measured {
		where += measured
	}
	if dim == "page" {
		where += ` AND page NOT LIKE '/post/%' AND page NOT LIKE '/en/post/%'`
	}
	q := `SELECT ` + col.expr + ` label, COUNT(*) n FROM Visits WHERE ` + where +
		` GROUP BY label ORDER BY n DESC LIMIT ?`
	rows, err := Db.Query(q, days, analytics.BotSQLPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LabelCount
	for rows.Next() {
		var lc LabelCount
		if err := rows.Scan(&lc.Label, &lc.Count); err != nil {
			return nil, err
		}
		out = append(out, lc)
	}
	return out, rows.Err()
}

// MeasuredSince devuelve la fecha de la primera vista medida en el navegador
// ("" si aún no hay ninguna). La UI la usa para explicar desde cuándo hay
// datos de duración, referrer y dispositivo.
func MeasuredSince() (string, error) {
	var d sql.NullString
	err := Db.QueryRow(`SELECT DATE_FORMAT(MIN(timestamp), '%Y-%m-%d') FROM Visits WHERE view_id IS NOT NULL`).Scan(&d)
	return d.String, err
}

// Today devuelve la fecha actual según la base de datos. Las fechas de las
// visitas se agrupan con la zona horaria de MySQL, que puede no ser la del
// servidor de la app; rellenar huecos con esta fecha evita desfases de un día.
func Today() (string, error) {
	var d string
	err := Db.QueryRow(`SELECT DATE_FORMAT(CURDATE(), '%Y-%m-%d')`).Scan(&d)
	return d, err
}

// ── Mantenimiento y privacidad ─────────────────────────────────────────────

// DeleteBotVisits borra visitas de bots (registradas por la medición antigua,
// que contaba desde el servidor). Devuelve cuántas filas borró.
func DeleteBotVisits() (int64, error) {
	res, err := Db.Exec(`DELETE FROM Visits WHERE REGEXP_LIKE(user_agent, ?, 'i') OR user_agent = ''`, analytics.BotSQLPattern)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteVisitsOlderThan aplica la política de retención.
func DeleteVisitsOlderThan(days int) (int64, error) {
	res, err := Db.Exec(`DELETE FROM Visits WHERE timestamp < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// AnonymizeLegacyIPs reemplaza las IPs reales guardadas antes de la medición
// anónima por su VisitorID. Una IP real contiene "." (IPv4) o ":" (IPv6); un
// VisitorID es hexadecimal, así que la operación es idempotente.
func AnonymizeLegacyIPs() (int, error) {
	total := 0
	for {
		rows, err := Db.Query(`SELECT DISTINCT ip FROM Visits WHERE ip LIKE '%.%' OR ip LIKE '%:%' LIMIT 500`)
		if err != nil {
			return total, err
		}
		var ips []string
		for rows.Next() {
			var ip string
			if err := rows.Scan(&ip); err != nil {
				rows.Close()
				return total, err
			}
			ips = append(ips, ip)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return total, err
		}
		if len(ips) == 0 {
			return total, nil
		}
		for _, ip := range ips {
			if _, err := Db.Exec(`UPDATE Visits SET ip = ? WHERE ip = ?`, analytics.VisitorID(ip), ip); err != nil {
				return total, err
			}
			total++
		}
	}
}

// UpdateVisitGeo completa la ubicación de una vista registrada por /api/track.
func UpdateVisitGeo(viewID, country, region, city string, lat, lon float64) error {
	_, err := Db.Exec(`UPDATE Visits SET country = ?, region = ?, city = ?, latitude = ?, longitude = ? WHERE view_id = ?`,
		country, region, city, lat, lon, viewID)
	return err
}
