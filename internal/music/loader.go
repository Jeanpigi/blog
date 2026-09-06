package music

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	mu         sync.Mutex
	MusicFiles []string
)

func LoadMusicFiles(folder string) error {
	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			mu.Lock()
			MusicFiles = append(MusicFiles, path)
			mu.Unlock()
		}
		return nil
	})
}

// AddFile agrega un archivo de forma segura para concurrencia: uploadHandler
// puede llamarlo desde una goroutine de request mientras el timer de la radio
// (en internal/handlers/streamHandler.go) lee la lista en otra goroutine.
func AddFile(path string) {
	mu.Lock()
	MusicFiles = append(MusicFiles, path)
	mu.Unlock()
}

// Snapshot devuelve una copia de la lista actual, segura para leer aunque
// otra goroutine esté agregando archivos concurrentemente vía AddFile.
func Snapshot() []string {
	mu.Lock()
	defer mu.Unlock()
	out := make([]string, len(MusicFiles))
	copy(out, MusicFiles)
	return out
}
