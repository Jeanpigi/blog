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

// renderErrorPage muestra una página de error personalizada
func renderErrorPage(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	utils.RenderTemplate(w, "templates/error.html", map[string]string{
		"Message": message,
	})
}

// UploadHandler gestiona la subida de archivos de música
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
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			renderErrorPage(w, "El archivo excede el límite de tamaño (10 MB).")
			return
		}

		// Obtener todos los archivos subidos
		files := r.MultipartForm.File["musicFiles"]
		if len(files) == 0 {
			renderErrorPage(w, "No se subió ningún archivo.")
			return
		}

		// Procesar cada archivo
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				renderErrorPage(w, "No se pudo leer el archivo.")
				return
			}
			defer file.Close()

			// Validar el tipo MIME
			if !isMusicFile(fileHeader) {
				renderErrorPage(w, "El archivo no es un MP3 válido.")
				return
			}

			// Sanitiza nombre de archivo y guarda con prefijo de timestamp
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
			dstPath := filepath.Join("./music", filename)
			dst, err := os.Create(dstPath)
			if err != nil {
				renderErrorPage(w, "Error al guardar el archivo.")
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				renderErrorPage(w, "Error al copiar el contenido del archivo.")
				return
			}

			// Agregar el archivo a la lista de música en memoria
			music.MusicFiles = append(music.MusicFiles, dst.Name())
		}

		// Redirigir para evitar reenvío del formulario
		http.Redirect(w, r, "/radio/upload", http.StatusSeeOther)
	}
}
