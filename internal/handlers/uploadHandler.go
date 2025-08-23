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
	"github.com/Jeanpigi/blog/internal/utils"
)

const (
	maxPerFile = 20 << 20 // 20 MB
	// Límite total del body por request (ej.: 5 archivos * 20MB + margen)
	maxRequestBody = 120 << 20 // 120 MB
)

var safeNameRx = regexp.MustCompile(`[^a-zA-Z0-9._\- ]+`)

// Validación mixta: header + extensión
func looksLikeMP3ByHeader(header *multipart.FileHeader) bool {
	mimeType := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))
	return mimeType == "audio/mpeg" && ext == ".mp3"
}

func detectRealMIME(f multipart.File) (string, error) {
	// Leer primeros 512 bytes
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	// Volver al inicio para no perder bytes al copiar
	if seeker, ok := f.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}
	return http.DetectContentType(buf[:n]), nil
}

func sanitizeName(name string) string {
	base := filepath.Base(name) // quita rutas
	base = strings.TrimSpace(base)
	base = safeNameRx.ReplaceAllString(base, "_")
	return base
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := struct{ Title string }{Title: "Subir Música"}
		utils.RenderTemplate(w, "templates/upload.html", data)
		return

	case http.MethodPost:
		// Límite duro al tamaño total del request
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

		// Límite de memoria para multipart (el resto va a disco temporal)
		if err := r.ParseMultipartForm(25 << 20); err != nil {
			http.Error(w, "El archivo excede el límite de tamaño (25MB en memoria).", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["musicFiles"]
		if len(files) == 0 {
			http.Error(w, "No se subió ningún archivo.", http.StatusBadRequest)
			return
		}
		if len(files) > 5 {
			http.Error(w, "Solo se permiten hasta 5 archivos.", http.StatusBadRequest)
			return
		}

		type SkipInfo struct {
			File   string `json:"file"`
			Reason string `json:"reason"`
		}
		var uploaded []string
		var skipped []SkipInfo

		// Asegura que exista la carpeta de destino
		dstDir := "./music"
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			http.Error(w, "No se pudo preparar el directorio de destino.", http.StatusInternalServerError)
			return
		}

		for _, fh := range files {
			name := sanitizeName(fh.Filename)

			// Tamaño por archivo
			if fh.Size > maxPerFile {
				skipped = append(skipped, SkipInfo{name, "Archivo demasiado grande (máx 20MB)"})
				continue
			}

			// Chequeo rápido por header/extensión
			if !looksLikeMP3ByHeader(fh) {
				skipped = append(skipped, SkipInfo{name, "Tipo declarado no permitido (solo MP3)"})
				continue
			}

			// Evita duplicados por nombre
			dstPath := filepath.Join(dstDir, name)
			if _, err := os.Stat(dstPath); err == nil {
				skipped = append(skipped, SkipInfo{name, "Ya existe en la carpeta"})
				continue
			}

			// Abrir archivo
			f, err := fh.Open()
			if err != nil {
				skipped = append(skipped, SkipInfo{name, "No se pudo abrir el archivo"})
				continue
			}

			// Detectar MIME real por contenido
			if real := func() string {
				mime, err := detectRealMIME(f)
				if err != nil {
					return ""
				}
				return mime
			}(); real != "" && real != "audio/mpeg" && real != "application/octet-stream" {
				// algunos MP3 antiguos caen en octet-stream; lo permitimos
				f.Close()
				skipped = append(skipped, SkipInfo{name, "Contenido no parece MP3"})
				continue
			}

			// Crear destino con permisos seguros
			dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				f.Close()
				skipped = append(skipped, SkipInfo{name, "Error al guardar el archivo"})
				continue
			}

			// Copiar contenido
			if _, err := io.Copy(dst, f); err != nil {
				skipped = append(skipped, SkipInfo{name, "Error al copiar el contenido"})
			} else {
				uploaded = append(uploaded, name)
				music.MusicFiles = append(music.MusicFiles, dstPath)
			}

			dst.Close()
			f.Close()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"uploaded": uploaded,
			"skipped":  skipped,
		})
		return

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}
