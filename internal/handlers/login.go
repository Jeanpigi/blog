package handlers

import (
	"github.com/Jeanpigi/blog/internal/utils"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Mostrar el formulario de inicio de sesión
		utils.RenderTemplate(w, "templates/login.html", nil)
	} else if r.Method == "POST" {
		// Procesar el formulario de inicio de sesión
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Verificar las credenciales y autenticar al usuario
		if utils.AuthenticateUser(username, password) {
			// Iniciar sesión y redirigir al dashboard
			utils.StartSession(w, r, username)
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		} else {
			utils.RenderTemplate(w, "templates/login.html", nil)
		}
	}
}
