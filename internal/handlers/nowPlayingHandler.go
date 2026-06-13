package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type nowPlayingResponse struct {
	Song      string `json:"song"`
	StartedAt int64  `json:"startedAt"` // Unix milliseconds
	ServerNow int64  `json:"serverNow"` // Unix ms del servidor: permite corregir el desfase del reloj del cliente
}

func NowPlayingHandler(w http.ResponseWriter, r *http.Request) {
	song, startedAt := GetBroadcastState()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(nowPlayingResponse{
		Song:      song,
		StartedAt: startedAt.UnixMilli(),
		ServerNow: time.Now().UnixMilli(),
	})
}

// AdvanceSongHandler skips to the next song immediately (admin endpoint).
func AdvanceSongHandler(w http.ResponseWriter, r *http.Request) {
	SkipBroadcast()
	NowPlayingHandler(w, r)
}
