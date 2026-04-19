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
	lastPlayed  string
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

// AddSong inserta una canción nueva en una posición aleatoria dentro de las canciones
// que aún no han sonado en el ciclo actual, para que suene pronto sin esperar al siguiente ciclo.
func AddSong(path string) {
	mu.Lock()
	defer mu.Unlock()

	// Si la playlist está vacía o ya terminó el ciclo, simplemente añadir al final.
	if len(pl) == 0 || currentSong >= len(pl) {
		pl = append(pl, path)
		return
	}

	// Insertar en posición aleatoria entre currentSong y el final.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pos := currentSong + r.Intn(len(pl)-currentSong+1)

	pl = append(pl, "")
	copy(pl[pos+1:], pl[pos:])
	pl[pos] = path
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
		// Evitar que la primera canción del nuevo ciclo sea igual a la última reproducida.
		if len(pl) > 1 && pl[0] == lastPlayed {
			pl[0], pl[1] = pl[1], pl[0]
		}
		currentSong = 0
	}
	song := pl[currentSong]
	lastPlayed = song
	currentSong++
	return song
}

func shuffle(s []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}
