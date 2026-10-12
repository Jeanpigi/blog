package site

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("SUPPORT_BREB_KEY", " @jbearp ")
	t.Setenv("SUPPORT_PAY_URL", "https://checkout.wompi.co/l/abc")
	t.Setenv("SUPPORT_PAYPAL_URL", "http://paypal.me/inseguro") // sin https: se ignora
	t.Setenv("CONTACT_EMAIL", "no-es-un-correo")                // inválido: se ignora
	t.Setenv("CONTACT_WHATSAPP", "+573001234567")

	l := Load()
	if l.BreBKey != "@jbearp" || l.PayURL == "" {
		t.Errorf("valores válidos no cargados: %+v", l)
	}
	if l.PayPalURL != "" || l.Email != "" {
		t.Errorf("valores inválidos deberían ignorarse: %+v", l)
	}
	if l.WhatsApp != "573001234567" {
		t.Errorf("WhatsApp = %q", l.WhatsApp)
	}
	if !l.SupportEnabled() || !l.ContactEnabled() {
		t.Error("SupportEnabled/ContactEnabled deberían ser true")
	}
}

func TestLoadEmpty(t *testing.T) {
	for _, k := range []string{"SUPPORT_BREB_KEY", "SUPPORT_PAY_URL", "SUPPORT_PAYPAL_URL", "CONTACT_EMAIL", "CONTACT_WHATSAPP"} {
		t.Setenv(k, "")
	}
	l := Load()
	if l.SupportEnabled() || l.ContactEnabled() {
		t.Errorf("sin configuración no debe haber métodos: %+v", l)
	}
}
