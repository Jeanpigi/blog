package music

import (
	"os"
	"path/filepath"
)

var MusicFiles []string

func LoadMusicFiles(folder string) error {
	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			MusicFiles = append(MusicFiles, path)
		}
		return nil
	})
}
