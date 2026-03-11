package playlist

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Jeanpigi/blog/internal/music"
)

var (
	pl          []string
	currentSong int
	mu          sync.Mutex
)

func CreatePlaylist() {
	mu.Lock()
	defer mu.Unlock()
	pl = make([]string, len(music.MusicFiles))
	copy(pl, music.MusicFiles)
	shuffle(pl)
	currentSong = 0
}

// AddSong agrega una canción nueva a la playlist activa sin reiniciar el servidor.
func AddSong(path string) {
	mu.Lock()
	defer mu.Unlock()
	pl = append(pl, path)
}

func NextSong() string {
	mu.Lock()
	defer mu.Unlock()

	// Si la playlist está vacía pero hay archivos cargados, crearla ahora.
	if len(pl) == 0 && len(music.MusicFiles) > 0 {
		pl = make([]string, len(music.MusicFiles))
		copy(pl, music.MusicFiles)
		shuffle(pl)
		currentSong = 0
	}

	if len(pl) == 0 {
		return ""
	}

	if currentSong >= len(pl) {
		shuffle(pl)
		currentSong = 0
	}
	song := pl[currentSong]
	currentSong++
	return song
}

func shuffle(s []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}
