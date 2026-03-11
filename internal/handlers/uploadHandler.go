package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Jeanpigi/blog/internal/music"
	"github.com/Jeanpigi/blog/internal/playlist"
	"github.com/Jeanpigi/blog/internal/utils"
)

const (
	maxPerFile     = 20 << 20  // 20 MB por archivo
	maxRequestBody = 120 << 20 // 120 MB total (5 archivos × ~20MB + margen)
)

var safeNameRx = regexp.MustCompile(`[^a-zA-Z0-9._\- ]+`)

// looksLikeMP3ByHeader valida extensión y Content-Type declarado.
func looksLikeMP3ByHeader(header *multipart.FileHeader) bool {
	mime := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))
	return ext == ".mp3" && (mime == "audio/mpeg" || mime == "audio/mp3" || mime == "application/octet-stream")
}

// isMagicMP3 verifica los magic bytes reales del archivo (ID3v2 o sync frame MPEG).
func isMagicMP3(buf []byte) bool {
	if len(buf) < 3 {
		return false
	}
	// ID3v2 tag header
	if buf[0] == 'I' && buf[1] == 'D' && buf[2] == '3' {
		return true
	}
	// MPEG audio sync word (0xFF seguido de 0xE0–0xFF)
	// Capa debe ser 01, 10 o 11 (no 00), y bits de audio válidos
	if buf[0] == 0xFF && (buf[1]&0xE0 == 0xE0) {
		layer := (buf[1] >> 1) & 0x03
		return layer != 0
	}
	return false
}

func readMagicBytes(f multipart.File) ([]byte, error) {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if seeker, ok := f.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}
	return buf[:n], nil
}

func sanitizeName(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSpace(base)
	base = safeNameRx.ReplaceAllString(base, "_")
	if base == "" || base == "." {
		return "unnamed.mp3"
	}
	return base
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.RenderTemplate(w, "templates/upload.html", struct{ Title string }{"Subir Música"})

	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "El request supera el límite de tamaño (120MB).", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["musicFiles"]
		if len(files) == 0 {
			http.Error(w, "No se subió ningún archivo.", http.StatusBadRequest)
			return
		}
		if len(files) > 5 {
			http.Error(w, "Solo se permiten hasta 5 archivos por request.", http.StatusBadRequest)
			return
		}

		type SkipInfo struct {
			File   string `json:"file"`
			Reason string `json:"reason"`
		}
		var uploaded []string
		var skipped []SkipInfo

		dstDir := "./music"
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			http.Error(w, "No se pudo preparar el directorio de destino.", http.StatusInternalServerError)
			return
		}

		for _, fh := range files {
			name := sanitizeName(fh.Filename)

			// 1. Tamaño por archivo
			if fh.Size > maxPerFile {
				skipped = append(skipped, SkipInfo{name, "Demasiado grande (máx 20MB)"})
				continue
			}

			// 2. Extensión + Content-Type declarado
			if !looksLikeMP3ByHeader(fh) {
				skipped = append(skipped, SkipInfo{name, "Solo se aceptan archivos .mp3"})
				continue
			}

			// 3. Evitar duplicados
			dstPath := filepath.Join(dstDir, name)
			if _, err := os.Stat(dstPath); err == nil {
				skipped = append(skipped, SkipInfo{name, "Ya existe en la carpeta"})
				continue
			}

			// 4. Abrir y verificar magic bytes reales
			f, err := fh.Open()
			if err != nil {
				skipped = append(skipped, SkipInfo{name, "No se pudo abrir el archivo"})
				continue
			}

			magic, err := readMagicBytes(f)
			if err != nil || !isMagicMP3(magic) {
				f.Close()
				skipped = append(skipped, SkipInfo{name, "El contenido no corresponde a un MP3 válido"})
				continue
			}

			// 5. Guardar en disco con permisos seguros
			dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				f.Close()
				skipped = append(skipped, SkipInfo{name, "Error al guardar el archivo"})
				continue
			}

			if _, err := io.Copy(dst, f); err != nil {
				dst.Close()
				f.Close()
				os.Remove(dstPath) // limpiar archivo parcial
				skipped = append(skipped, SkipInfo{name, "Error al escribir el archivo"})
				continue
			}

			dst.Close()
			f.Close()

			// 6. Actualizar música y playlist en memoria (sin reiniciar servidor)
			music.MusicFiles = append(music.MusicFiles, dstPath)
			playlist.AddSong(dstPath)
			uploaded = append(uploaded, name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"uploaded": uploaded,
			"skipped":  skipped,
		})

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
