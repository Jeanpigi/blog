package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		remote  string
		want    string
	}{
		{"nginx reenvía la IP verificada", map[string]string{"X-Real-IP": "190.1.2.3"}, "127.0.0.1:5000", "190.1.2.3"},
		{"CF-Connecting-IP falsificado se ignora", map[string]string{"X-Real-IP": "200.9.9.9", "CF-Connecting-IP": "1.1.1.1"}, "127.0.0.1:5000", "200.9.9.9"},
		{"sin nginx: IP de la conexión", nil, "10.0.0.5:4321", "10.0.0.5"},
		{"IPv6 con puerto", nil, "[2001:db8::1]:443", "2001:db8::1"},
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = c.remote
		for k, v := range c.headers {
			r.Header.Set(k, v)
		}
		if got := ClientIP(r); got != c.want {
			t.Errorf("%s: ClientIP = %q, want %q", c.name, got, c.want)
		}
	}
}
