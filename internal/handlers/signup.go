package handlers

import (
	"log"
	"net/http"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/models"
	"github.com/Jeanpigi/blog/internal/utils"
)

// SignupHandler — GET/POST /signup. Solo un usuario autenticado puede crear
// cuentas nuevas (el router lo envuelve en RequireAuth). La sesión actual NO
// cambia al usuario creado: el admin sigue logueado como sí mismo.
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		csrfToken := utils.IssueCSRFToken(w)
		utils.RenderTemplate(w, r, "templates/signup.html", map[string]interface{}{
			"CsrfToken": csrfToken,
		})
		return
	}

	fail := func(code string) {
		http.Redirect(w, r, "/signup?error="+code, http.StatusSeeOther)
	}

	if !utils.ValidateCSRFForm(r) {
		fail("csrf_invalid")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if !validUsername.MatchString(username) {
		fail("invalid_username")
		return
	}
	if len(password) < 8 {
		fail("weak_password")
		return
	}
	if password != r.FormValue("password_confirm") {
		fail("password_mismatch")
		return
	}

	existingUser, err := db.GetUserByUsername(username)
	if err != nil {
		log.Println("❌ Error consultando usuario:", err)
		fail("server_error")
		return
	}
	if existingUser != nil {
		fail("user_exists")
		return
	}

	newUser := models.User{
		Username: username,
		Password: utils.HashPassword(password),
	}
	if err := db.InsertUser(&newUser); err != nil {
		log.Println("❌ Error al insertar el usuario:", err)
		fail("server_error")
		return
	}

	http.Redirect(w, r, "/dashboard?ok=user_created", http.StatusSeeOther)
}
