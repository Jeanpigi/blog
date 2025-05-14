package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Jeanpigi/blog/internal/music"
	"github.com/Jeanpigi/blog/internal/utils"
)

// isMusicFile verifica si el archivo tiene el MIME type adecuado.
func isMusicFile(header *multipart.FileHeader) bool {
	mimeType := header.Header.Get("Content-Type")
	return mimeType == "audio/mpeg"
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := struct {
			Title string
		}{
			Title: "Upload Music",
		}
		utils.RenderTemplate(w, "templates/upload.html", data)
		return
	}

	if r.Method == http.MethodPost {
		// Limita el tamaño del formulario a 10MB
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error processing form data", http.StatusBadRequest)
			return
		}

		// Obtener todos los archivos subidos con la key "musicFiles"
		files := r.MultipartForm.File["musicFiles"]
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			return
		}

		// Itera sobre cada archivo
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error retrieving file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Valida el tipo MIME
			if !isMusicFile(fileHeader) {
				http.Error(w, "Invalid file type", http.StatusBadRequest)
				return
			}

			// Sanitiza el nombre y agrega un prefijo con timestamp
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
			dstPath := filepath.Join("./music", filename)
			dst, err := os.Create(dstPath)
			if err != nil {
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}

			// Agrega la ruta del archivo a la lista en memoria
			music.MusicFiles = append(music.MusicFiles, dst.Name())
		}

		// Redirige a la misma página después del POST para evitar reenvío de formulario
		http.Redirect(w, r, "/radio/upload", http.StatusSeeOther)
	}
}
