package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Jeanpigi/blog/internal/playlist"
)

// nowPlaying mantiene la canción actual para que requests de rango consecutivos
// sirvan el mismo archivo (un request sin Range = nueva canción).
var (
	nowPlaying   string
	nowPlayingMu sync.Mutex
)

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	rangeHeader := r.Header.Get("Range")

	nowPlayingMu.Lock()
	if rangeHeader == "" || nowPlaying == "" {
		nowPlaying = playlist.NextSong()
	}
	songPath := nowPlaying
	nowPlayingMu.Unlock()

	if songPath == "" {
		http.Error(w, "No hay canciones en la playlist. Sube archivos MP3 primero.", http.StatusServiceUnavailable)
		return
	}

	file, err := os.Open(songPath)
	if err != nil {
		// El archivo podría haber sido borrado; avanzar a la siguiente canción
		nowPlayingMu.Lock()
		nowPlaying = ""
		nowPlayingMu.Unlock()
		http.Error(w, "Canción no disponible", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	// Headers de radio/audio
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Nombre de la canción actual (útil para reproductores)
	w.Header().Set("X-Song-Name", filepath.Base(songPath))

	// http.ServeContent maneja automáticamente:
	// - Range requests (206 Partial Content)
	// - ETag y Last-Modified
	// - Content-Length correcto
	// - Buffer interno optimizado (~32KB, adecuado para 128kbps ≈ 16KB/s)
	http.ServeContent(w, r, filepath.Base(songPath), stat.ModTime(), file)
}
