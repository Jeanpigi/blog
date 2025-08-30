package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	db "github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
)

// ===== DTO liviano para listados =====
type PostListItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Categoria   string `json:"categoria"`
	ReadingMin  int    `json:"reading_min"` // ⬅️ NUEVO
}

// GET /api/posts  (acepta ?limit=&offset= y ?full=1)
func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseInt(r.URL.Query().Get("limit"), 0), 0, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))
	full := parseBool(r.URL.Query().Get("full"))

	var posts []*models.Post
	var err error
	if limit > 0 {
		posts, err = db.GetPostsPaged(limit, offset)
	} else {
		posts, err = db.GetAllPosts()
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if full {
		_ = json.NewEncoder(w).Encode(posts) // blog “full”
		return
	}

	items := toListDTO(posts) // listados livianos con reading_min
	_ = json.NewEncoder(w).Encode(items)
}

// ===== helpers compartidos =====

func toListDTO(posts []*models.Post) []PostListItem {
	items := make([]PostListItem, 0, len(posts))
	for _, p := range posts {
		mins := readingMinutesFromHTML(string(p.Content)) // ⬅️ calcula con Content completo
		items = append(items, PostListItem{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			Categoria:   p.Categoria,
			ReadingMin:  mins,
		})
	}
	return items
}

func readingMinutesFromHTML(html string) int {
	plain := stripTags(html)
	words := countWords(plain)
	mins := (words + 199) / 200 // 200 wpm
	if mins < 1 {
		mins = 1
	}
	return mins
}

func stripTags(s string) string {
	in := false
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '<':
			in = true
		case '>':
			in = false
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func countWords(s string) int {
	f := strings.Fields(strings.TrimSpace(s))
	return len(f)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}
