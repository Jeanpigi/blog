package analytics

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// GeoInfo representa la respuesta de ip-api.com.
type GeoInfo struct {
	Country   string  `json:"country"`
	Region    string  `json:"regionName"`
	City      string  `json:"city"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type geoCacheEntry struct {
	info    GeoInfo
	expires time.Time
}

var (
	geoCache  sync.Map
	geoClient = &http.Client{Timeout: 3 * time.Second}
)

// LookupGeo geolocaliza una IP (ip-api.com, gratis solo por HTTP). Cachea 24h
// por IP. Es bloqueante: llámalo desde una goroutine.
func LookupGeo(ip string) GeoInfo {
	if v, ok := geoCache.Load(ip); ok {
		entry := v.(geoCacheEntry)
		if time.Now().Before(entry.expires) {
			return entry.info
		}
		geoCache.Delete(ip)
	}

	var info GeoInfo
	resp, err := geoClient.Get("http://ip-api.com/json/" + url.PathEscape(ip) + "?fields=country,regionName,city,lat,lon")
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Println("⚠️ Error decodificando geo:", err)
		return info
	}

	geoCache.Store(ip, geoCacheEntry{info: info, expires: time.Now().Add(24 * time.Hour)})
	return info
}
