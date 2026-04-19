package handlers

import (
	"encoding/json"
	"net/http"
)

type nowPlayingResponse struct {
	Song      string `json:"song"`
	StartedAt int64  `json:"startedAt"` // Unix milisegundos
}

func NowPlayingHandler(w http.ResponseWriter, r *http.Request) {
	song, startedAt := GetBroadcastState()
	// Si no hay canción activa (playlist vacía al inicio), intentar obtener una ahora.
	if song == "" {
		AdvanceBroadcast()
		song, startedAt = GetBroadcastState()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nowPlayingResponse{
		Song:      song,
		StartedAt: startedAt.UnixMilli(),
	})
}

// AdvanceSongHandler lo llaman los clientes cuando su audio termina.
// Avanza la emisión global a la siguiente canción.
func AdvanceSongHandler(w http.ResponseWriter, r *http.Request) {
	AdvanceBroadcast()
	NowPlayingHandler(w, r)
}
