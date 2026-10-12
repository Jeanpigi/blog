package analytics

import (
	"testing"
	"time"
)

func TestIsBot(t *testing.T) {
	bots := []string{
		"", "Googlebot/2.1 (+http://www.google.com/bot.html)", "curl/8.4.0",
		"Mozilla/5.0 (compatible; AhrefsBot/7.0)", "python-requests/2.31", "facebookexternalhit/1.1",
	}
	for _, ua := range bots {
		if !IsBot(ua) {
			t.Errorf("IsBot(%q) = false, se esperaba true", ua)
		}
	}
	human := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Safari/537.36"
	if IsBot(human) {
		t.Errorf("IsBot(navegador) = true")
	}
}

func TestDevice(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile/15E148":       "mobile",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8) Chrome/128.0 Mobile Safari/537.36": "mobile",
		"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)":                              "tablet",
		"Mozilla/5.0 (Linux; Android 13; SM-X200) Chrome/128.0 Safari/537.36":        "tablet",
		"Mozilla/5.0 (X11; Linux x86_64) Firefox/131.0":                              "desktop",
	}
	for ua, want := range cases {
		if got := Device(ua); got != want {
			t.Errorf("Device(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestReferrerHost(t *testing.T) {
	own := "blog.jbearp.com"
	cases := map[string]string{
		"":                                       "",
		"https://blog.jbearp.com/blog":           "",
		"https://www.google.com.co/":             "google.com",
		"https://www.google.com/search?q=go":     "google.com",
		"https://l.instagram.com/?u=x":           "instagram.com",
		"https://t.co/abc":                       "x.com",
		"https://news.ycombinator.com/item?id=1": "news.ycombinator.com",
		"no es una url":                          "",
	}
	for ref, want := range cases {
		if got := ReferrerHost(ref, own); got != want {
			t.Errorf("ReferrerHost(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestTrackablePath(t *testing.T) {
	for _, p := range []string{"/", "/blog", "/historias", "/tecnologias", "/portafolio", "/post/42", "/en", "/en/blog", "/en/post/42"} {
		if !TrackablePath(p) {
			t.Errorf("TrackablePath(%q) = false", p)
		}
	}
	for _, p := range []string{"/dashboard", "/visitas", "/login", "/post/abc", "/post/1/x", "/api/posts", "/en/dashboard", "/english"} {
		if TrackablePath(p) {
			t.Errorf("TrackablePath(%q) = true", p)
		}
	}
}

func TestVisitorID(t *testing.T) {
	t.Setenv("VISITOR_HASH_KEY", "clave-de-prueba")
	a, b := VisitorID("1.2.3.4"), VisitorID("1.2.3.4")
	if a != b || len(a) != 32 {
		t.Fatalf("VisitorID no es estable o no mide 32: %q %q", a, b)
	}
	if a == VisitorID("1.2.3.5") {
		t.Error("IPs distintas dieron el mismo VisitorID")
	}
}

func TestLimiter(t *testing.T) {
	l := NewLimiter(2, time.Minute)
	for i := 0; i < 2; i++ {
		if !l.Allow("ip") {
			t.Fatalf("el evento %d debería pasar", i+1)
		}
	}
	if l.Allow("ip") {
		t.Error("el tercer evento debería bloquearse")
	}
	if !l.Allow("otra-ip") {
		t.Error("el límite es por clave")
	}
}
