package handlers

import (
	"encoding/json"
	"net/http"
)

type nowPlayingResponse struct {
	Song      string `json:"song"`
	StartedAt int64  `json:"startedAt"` // Unix milliseconds
}

func NowPlayingHandler(w http.ResponseWriter, r *http.Request) {
	song, startedAt := GetBroadcastState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nowPlayingResponse{
		Song:      song,
		StartedAt: startedAt.UnixMilli(),
	})
}

// AdvanceSongHandler skips to the next song immediately (admin endpoint).
func AdvanceSongHandler(w http.ResponseWriter, r *http.Request) {
	SkipBroadcast()
	NowPlayingHandler(w, r)
}
