package playlist

import (
	"math/rand"
	"time"

	"github.com/Jeanpigi/blog/internal/music"
)

var (
	Playlist    []string
	CurrentSong int
)

func CreatePlaylist() {
	Playlist = make([]string, len(music.MusicFiles))
	copy(Playlist, music.MusicFiles)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(Playlist), func(i, j int) {
		Playlist[i], Playlist[j] = Playlist[j], Playlist[i]
	})
}

func NextSong() string {
	if CurrentSong >= len(Playlist) {
		CreatePlaylist() // Recreate the playlist once all songs have been played
		CurrentSong = 0
	}
	song := Playlist[CurrentSong]
	CurrentSong++
	return song
}
