package models

import "time"

type Visit struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"` // VisitorID (hash HMAC de la IP), nunca la IP real
	UserAgent string    `json:"user_agent"`
	Page      string    `json:"page"`
	Country   string    `json:"country"`
	Region    string    `json:"region"`
	City      string    `json:"city"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	ISP       string    `json:"isp"`
	// Medición client-side (ver /api/track):
	ViewID      string `json:"-"`        // id aleatorio de la vista, para reportar su duración
	Referrer    string `json:"referrer"` // dominio externo de origen ("" = directo/interno)
	Device      string `json:"device"`   // mobile | tablet | desktop
	DurationSec int    `json:"duration_sec"`
}
