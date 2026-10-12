// Package site contiene la configuración pública del sitio que viene del
// .env: medios de apoyo (Bre-B, link de pago, PayPal) y de contacto.
package site

import (
	"os"
	"regexp"
	"strings"
)

// Links son los datos de contacto y apoyo, configurables en el .env. Cada
// método aparece en la web solo si está configurado, así se pueden ir
// activando a medida que se abren las cuentas.
type Links struct {
	BreBKey   string // llave Bre-B (ej. "@jbearp"): Nequi, Daviplata, bancos
	PayURL    string // link de pago (Wompi, Bold, Mercado Pago…): PSE, Nequi, tarjetas
	PayPalURL string // PayPal para apoyos desde el exterior
	Email     string // correo de contacto para servicios
	WhatsApp  string // número internacional sin "+" (ej. 573001234567)
}

var (
	waRx    = regexp.MustCompile(`^[0-9]{8,15}$`)
	emailRx = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

// LoadLinks lee y valida la configuración. Un valor inválido se ignora
// (mejor no mostrar el método que mostrar un enlace roto).
func Load() Links {
	httpsOnly := func(u string) string {
		if strings.HasPrefix(u, "https://") {
			return u
		}
		return ""
	}
	l := Links{
		BreBKey:   strings.TrimSpace(os.Getenv("SUPPORT_BREB_KEY")),
		PayURL:    httpsOnly(strings.TrimSpace(os.Getenv("SUPPORT_PAY_URL"))),
		PayPalURL: httpsOnly(strings.TrimSpace(os.Getenv("SUPPORT_PAYPAL_URL"))),
		Email:     strings.TrimSpace(os.Getenv("CONTACT_EMAIL")),
		WhatsApp:  strings.TrimPrefix(strings.TrimSpace(os.Getenv("CONTACT_WHATSAPP")), "+"),
	}
	if len(l.BreBKey) > 60 {
		l.BreBKey = ""
	}
	if !emailRx.MatchString(l.Email) {
		l.Email = ""
	}
	if !waRx.MatchString(l.WhatsApp) {
		l.WhatsApp = ""
	}
	return l
}

// SupportEnabled indica si hay al menos un método de apoyo configurado.
func (l Links) SupportEnabled() bool {
	return l.BreBKey != "" || l.PayURL != "" || l.PayPalURL != ""
}

// ContactEnabled indica si hay al menos un medio de contacto para servicios.
func (l Links) ContactEnabled() bool {
	return l.Email != "" || l.WhatsApp != ""
}
