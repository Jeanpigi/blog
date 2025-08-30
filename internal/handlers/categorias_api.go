package handlers

import (
	"encoding/json"
	"net/http"

	db "github.com/Jeanpigi/blog/db"
)

// arriba del todo, fuera de las funciones:
const defaultPageSize = 12

// GET /api/categories?category=Tech&limit=12&offset=0
// GET /api/tech?limit=12&offset=0
func GetPostsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseInt(r.URL.Query().Get("limit"), defaultPageSize), 1, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))

	posts, err := db.FindPostsByCategoryPaged("Tech", limit, offset)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	items := toListDTO(posts) // usa el DTO liviano (sin Content)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// GET /api/histories?limit=&offset=
func GetPostsByHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// ⬇️ usa 12 por defecto si no viene
	limit := clamp(parseInt(r.URL.Query().Get("limit"), defaultPageSize), 1, 50)
	offset := max0(parseInt(r.URL.Query().Get("offset"), 0))

	posts, err := db.FindPostsByCategoryPaged("Historias", limit, offset)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	items := toListDTO(posts)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
