package handlers

import (
	"encoding/json"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
)

const defaultPageSize = 12

// GET /api/categories?limit=12&offset=0&paginated=1
func GetPostsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseInt(r.URL.Query().Get("limit"), defaultPageSize), 1, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))
	paginated := parseBool(r.URL.Query().Get("paginated"))

	items, err := db.FindPostsByCategoryListPaged("Tech", limit, offset)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if paginated {
		total, _ := db.CountPostsByCategory("Tech")
		_ = json.NewEncoder(w).Encode(PaginatedResponse{Items: items, Total: total, Limit: limit, Offset: offset})
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// GET /api/histories?limit=12&offset=0&paginated=1
func GetPostsByHistoryHandler(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseInt(r.URL.Query().Get("limit"), defaultPageSize), 1, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))
	paginated := parseBool(r.URL.Query().Get("paginated"))

	items, err := db.FindPostsByCategoryListPaged("Historias", limit, offset)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if paginated {
		total, _ := db.CountPostsByCategory("Historias")
		_ = json.NewEncoder(w).Encode(PaginatedResponse{Items: items, Total: total, Limit: limit, Offset: offset})
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}
