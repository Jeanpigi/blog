package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Jeanpigi/blog/internal/playlist"
)

// Una estructura para representar un rango
type httpRange struct {
	start, length int64
}

// parseRange toma el header Range y el tamaño total del archivo, y devuelve los rangos solicitados.
func parseRange(s string, size int64) ([]httpRange, error) {
	if !strings.HasPrefix(s, "bytes=") {
		return nil, fmt.Errorf("invalid range")
	}
	var ranges []httpRange
	noOverlap := false
	parts := strings.Split(s[len("bytes="):], ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		segs := strings.Split(part, "-")
		if len(segs) != 2 {
			return nil, fmt.Errorf("invalid range segment")
		}
		var r httpRange
		if segs[0] == "" {
			// Sufijo: -X significa los últimos X bytes
			suffix, err := strconv.ParseInt(segs[1], 10, 64)
			if err != nil {
				return nil, err
			}
			if suffix > size {
				suffix = size
			}
			r.start = size - suffix
			r.length = suffix
		} else {
			// Rango normal: X-Y
			start, err := strconv.ParseInt(segs[0], 10, 64)
			if err != nil || start < 0 {
				return nil, fmt.Errorf("invalid start")
			}
			var end int64
			if segs[1] == "" {
				end = size - 1
			} else {
				end, err = strconv.ParseInt(segs[1], 10, 64)
				if err != nil || start > end {
					return nil, fmt.Errorf("invalid end")
				}
			}
			if start >= size {
				// Rango fuera del tamaño del archivo
				noOverlap = true
				continue
			}
			if end >= size {
				end = size - 1
			}
			r.start = start
			r.length = end - start + 1
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		return nil, fmt.Errorf("no overlap")
	}
	return ranges, nil
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	// Supongamos que songPath se obtiene de tu playlist
	songPath := playlist.NextSong()
	file, err := os.Open(songPath)
	if err != nil {
		http.Error(w, "Error al abrir el archivo", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error al obtener info del archivo", http.StatusInternalServerError)
		return
	}

	// Configura headers comunes
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "no-cache")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// Si no se solicita un rango, enviamos el archivo completo
		w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		io.Copy(w, file)
		return
	}

	// Parsear el header Range
	ranges, err := parseRange(rangeHeader, stat.Size())
	if err != nil || len(ranges) == 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", stat.Size()))
		http.Error(w, "Requested Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if len(ranges) == 1 {
		// Caso simple: un solo rango
		rng := ranges[0]
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rng.start, rng.start+rng.length-1, stat.Size()))
		w.Header().Set("Content-Length", strconv.FormatInt(rng.length, 10))
		w.WriteHeader(http.StatusPartialContent)
		file.Seek(rng.start, io.SeekStart)
		io.CopyN(w, file, rng.length)
		return
	}

	// Múltiples rangos: responder con multipart/byteranges
	boundary := "MY_BOUNDARY_123456" // Podrías generar uno aleatorio
	w.Header().Set("Content-Type", "multipart/byteranges; boundary="+boundary)
	w.WriteHeader(http.StatusPartialContent)

	for _, rng := range ranges {
		// Escribe la separación y headers para cada parte
		fmt.Fprintf(w, "--%s\r\n", boundary)
		fmt.Fprintf(w, "Content-Type: audio/mpeg\r\n")
		fmt.Fprintf(w, "Content-Range: bytes %d-%d/%d\r\n\r\n", rng.start, rng.start+rng.length-1, stat.Size())
		file.Seek(rng.start, io.SeekStart)
		io.CopyN(w, file, rng.length)
		fmt.Fprintf(w, "\r\n")
	}
	fmt.Fprintf(w, "--%s--\r\n", boundary)
}
