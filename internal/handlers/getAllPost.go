package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	db "github.com/Jeanpigi/blog/db"
)

// PaginatedResponse envuelve los items con metadatos de paginación.
type PaginatedResponse struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// GET /api/posts?limit=12&offset=0&paginated=1
func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseInt(r.URL.Query().Get("limit"), 0), 0, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))
	paginated := parseBool(r.URL.Query().Get("paginated"))

	w.Header().Set("Content-Type", "application/json")

	if limit > 0 {
		items, err := db.GetPostsListPaged(limit, offset)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if paginated {
			total, _ := db.CountPosts()
			_ = json.NewEncoder(w).Encode(PaginatedResponse{Items: items, Total: total, Limit: limit, Offset: offset})
		} else {
			_ = json.NewEncoder(w).Encode(items)
		}
		return
	}

	items, err := db.GetAllPostsList()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// ===== helpers compartidos =====

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
