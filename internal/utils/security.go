package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"sync"

	"github.com/Jeanpigi/blog/db"
	"golang.org/x/crypto/bcrypt"
)

var csrfTokens = make(map[string]string)
var csrfMutex sync.Mutex

// 🔐 Genera un token CSRF aleatorio
func GenerateCSRFToken(sessionID string) string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Error al generar token CSRF:", err)
		return ""
	}

	token := base64.StdEncoding.EncodeToString(b)

	csrfMutex.Lock()
	csrfTokens[sessionID] = token
	csrfMutex.Unlock()

	return token
}

// ✅ Valida el token CSRF recibido
func ValidateCSRF(sessionID, token string) bool {
	csrfMutex.Lock()
	defer csrfMutex.Unlock()

	expectedToken, exists := csrfTokens[sessionID]
	if !exists || expectedToken != token {
		return false
	}

	// Eliminar el token usado para evitar reutilización
	delete(csrfTokens, sessionID)
	return true
}

// 🔑 Hashea contraseñas de manera segura con bcrypt
func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal("Error al generar hash de contraseña:", err)
	}
	return string(hashedPassword)
}

// 🔓 Autenticación segura de usuarios
func AuthenticateUser(username, password string) bool {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		log.Println("❌ Error al obtener el usuario:", err)
		return false
	}
	if user == nil {
		log.Println("⚠️ Usuario no encontrado:", username)
		return false
	}

	// Comparar la contraseña hasheada almacenada con la ingresada
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("⚠️ Error al verificar la contraseña:", err)
		return false
	}

	log.Println("✅ Usuario autenticado correctamente:", username)
	return true
}




