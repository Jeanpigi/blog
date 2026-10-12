package analytics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"sync"
)

var (
	hashKey     []byte
	hashKeyOnce sync.Once
)

// key devuelve la clave del hash de visitantes: VISITOR_HASH_KEY si existe;
// si no, una derivada de SESSION_AUTH_KEY (así funciona sin configurar nada).
// Cambiar la clave hace que los visitantes "únicos" se cuenten de nuevo.
func key() []byte {
	hashKeyOnce.Do(func() {
		if k := os.Getenv("VISITOR_HASH_KEY"); k != "" {
			hashKey = []byte(k)
			return
		}
		m := hmac.New(sha256.New, []byte(os.Getenv("SESSION_AUTH_KEY")))
		m.Write([]byte("visitor-hash-v1"))
		hashKey = m.Sum(nil)
	})
	return hashKey
}

// VisitorID convierte una IP en un identificador anónimo y estable (32 hex).
// Permite contar visitantes únicos sin guardar la IP real; con HMAC y clave
// secreta no se puede revertir probando todas las IPs posibles.
func VisitorID(ip string) string {
	m := hmac.New(sha256.New, key())
	m.Write([]byte(ip))
	return hex.EncodeToString(m.Sum(nil))[:32]
}

// RetentionDays es cuántos días se conserva el detalle de cada visita
// (VISITS_RETENTION_DAYS, por defecto 400: cubre la vista de 12 meses).
func RetentionDays() int {
	if n, err := strconv.Atoi(os.Getenv("VISITS_RETENTION_DAYS")); err == nil && n >= 30 {
		return n
	}
	return 400
}
