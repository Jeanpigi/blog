package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Jeanpigi/blog/internal/playlist"
)

// broadcast mantiene el estado global de la emisión de radio.
// Todos los clientes reciben la misma canción en la misma posición.
var broadcast struct {
	song         string
	startedAt    time.Time
	lastAdvanced time.Time
	mu           sync.RWMutex
}

// InitBroadcast debe llamarse una vez después de que la playlist esté cargada.
func InitBroadcast() {
	broadcast.mu.Lock()
	defer broadcast.mu.Unlock()
	broadcast.song = playlist.NextSong()
	broadcast.startedAt = time.Now()
	broadcast.lastAdvanced = time.Now()
}

// GetBroadcastState devuelve el nombre base de la canción actual y cuándo empezó.
func GetBroadcastState() (string, time.Time) {
	broadcast.mu.RLock()
	defer broadcast.mu.RUnlock()
	return filepath.Base(broadcast.song), broadcast.startedAt
}

// AdvanceBroadcast avanza a la siguiente canción.
// Debounceado a 5s para evitar saltos cuando varios clientes terminan simultáneamente.
// Si no hay canción activa (broadcast vacío), omite el debounce para inicializar.
func AdvanceBroadcast() {
	broadcast.mu.Lock()
	defer broadcast.mu.Unlock()
	if broadcast.song != "" && time.Since(broadcast.lastAdvanced) < 5*time.Second {
		return
	}
	next := playlist.NextSong()
	if next == "" {
		return
	}
	broadcast.song = next
	broadcast.startedAt = time.Now()
	broadcast.lastAdvanced = time.Now()
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	broadcast.mu.RLock()
	songPath := broadcast.song
	broadcast.mu.RUnlock()

	// Si no hay canción activa (playlist vacía al inicio y canciones subidas después),
	// intentar obtener una ahora.
	if songPath == "" {
		AdvanceBroadcast()
		broadcast.mu.RLock()
		songPath = broadcast.song
		broadcast.mu.RUnlock()
	}

	if songPath == "" {
		http.Error(w, "No hay canciones disponibles. Sube archivos MP3 primero.", http.StatusServiceUnavailable)
		return
	}

	file, err := os.Open(songPath)
	if err != nil {
		http.Error(w, "Canción no disponible", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Song-Name", filepath.Base(songPath))
	w.Header().Set("Access-Control-Expose-Headers", "X-Song-Name")

	http.ServeContent(w, r, filepath.Base(songPath), stat.ModTime(), file)
}
