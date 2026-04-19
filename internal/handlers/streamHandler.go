package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Jeanpigi/blog/internal/music"
	"github.com/Jeanpigi/blog/internal/playlist"
)

var broadcast struct {
	song      string
	startedAt time.Time
	mu        sync.RWMutex
}

var (
	broadcastTimer   *time.Timer
	broadcastTimerMu sync.Mutex
)

// InitBroadcast must be called once after the playlist is loaded.
// Starts the server-side loop that auto-advances songs based on their duration.
func InitBroadcast() {
	next := playlist.NextSong()
	broadcast.mu.Lock()
	broadcast.song = next
	broadcast.startedAt = time.Now()
	broadcast.mu.Unlock()

	if next != "" {
		log.Printf("📻 Broadcast iniciado: %s", filepath.Base(next))
	}
	scheduleBroadcastAdvance(next)
}

// GetBroadcastState returns the current song base name and when it started.
func GetBroadcastState() (string, time.Time) {
	broadcast.mu.RLock()
	defer broadcast.mu.RUnlock()
	return filepath.Base(broadcast.song), broadcast.startedAt
}

// SkipBroadcast advances immediately to the next song (admin use).
func SkipBroadcast() {
	advanceBroadcast()
}

// StartBroadcastIfEmpty activates the broadcast only when no song is playing.
// Called after a new file is uploaded while the server had no music.
func StartBroadcastIfEmpty() {
	broadcast.mu.RLock()
	empty := broadcast.song == ""
	broadcast.mu.RUnlock()
	if empty {
		advanceBroadcast()
	}
}

func scheduleBroadcastAdvance(song string) {
	var dur time.Duration
	if song == "" {
		dur = 5 * time.Second
	} else {
		secs := music.Duration(song)
		if secs < 10 {
			secs = 240
		}
		dur = time.Duration(secs * float64(time.Second))
		log.Printf("📻 Siguiente avance en %.0fs (%s)", secs, filepath.Base(song))
	}

	broadcastTimerMu.Lock()
	defer broadcastTimerMu.Unlock()
	if broadcastTimer != nil {
		broadcastTimer.Stop()
	}
	broadcastTimer = time.AfterFunc(dur, advanceBroadcast)
}

func advanceBroadcast() {
	next := playlist.NextSong()
	if next == "" {
		scheduleBroadcastAdvance("")
		return
	}
	broadcast.mu.Lock()
	broadcast.song = next
	broadcast.startedAt = time.Now()
	broadcast.mu.Unlock()
	log.Printf("📻 Avanzando a: %s", filepath.Base(next))
	scheduleBroadcastAdvance(next)
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	broadcast.mu.RLock()
	songPath := broadcast.song
	broadcast.mu.RUnlock()

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
