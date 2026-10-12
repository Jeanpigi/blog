package handlers

import (
	"log"
	"time"

	"github.com/Jeanpigi/blog/db"
	"github.com/Jeanpigi/blog/internal/analytics"
)

// StartVisitsMaintenance limpia la tabla Visits al arrancar y luego una vez al
// día: borra bots heredados, anonimiza IPs antiguas y aplica la retención.
func StartVisitsMaintenance() {
	go func() {
		for {
			runVisitsMaintenance()
			time.Sleep(24 * time.Hour)
		}
	}()
}

func runVisitsMaintenance() {
	if n, err := db.DeleteBotVisits(); err != nil {
		log.Println("⚠️ mantenimiento visitas (bots):", err)
	} else if n > 0 {
		log.Printf("🧹 Visitas de bots borradas: %d", n)
	}
	if n, err := db.AnonymizeLegacyIPs(); err != nil {
		log.Println("⚠️ mantenimiento visitas (anonimizar):", err)
	} else if n > 0 {
		log.Printf("🔒 IPs anonimizadas: %d", n)
	}
	days := analytics.RetentionDays()
	if n, err := db.DeleteVisitsOlderThan(days); err != nil {
		log.Println("⚠️ mantenimiento visitas (retención):", err)
	} else if n > 0 {
		log.Printf("🗑️ Visitas de más de %d días borradas: %d", days, n)
	}
}
