package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rejeki-baru-farm/internal/config"
	"rejeki-baru-farm/internal/services"
)

func setup() http.Handler {
	services.InitAuth(services.AuthOptions{Secret: strings.Repeat("k", 40), TTL: time.Hour})
	return Setup(config.Config{Env: "development", AllowedOrigins: []string{"http://localhost:5173"}})
}

func do(h http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestTanpaLoginDitolak(t *testing.T) {
	h := setup()
	for _, p := range []string{"/api/me", "/api/sapi", "/api/kas", "/api/bo/export", "/api/users"} {
		if w := do(h, "GET", p, "", nil); w.Code != 401 {
			t.Errorf("GET %s tanpa login = %d, mau 401", p, w.Code)
		}
	}
	w := do(h, "GET", "/api/sapi", "", map[string]string{"Cookie": "rbf_session=token-palsu"})
	if w.Code != 401 {
		t.Errorf("cookie palsu = %d, mau 401", w.Code)
	}
}

func TestHeaderKeamanan(t *testing.T) {
	w := do(setup(), "GET", "/api/me", "", nil)
	for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Referrer-Policy", "Cache-Control"} {
		if w.Header().Get(k) == "" {
			t.Errorf("header %s tidak ada", k)
		}
	}
}

func TestCORS(t *testing.T) {
	h := setup()
	w := do(h, "OPTIONS", "/api/login", "", map[string]string{"Origin": "https://jahat.example", "Access-Control-Request-Method": "POST"})
	if w.Code != 403 || w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("origin asing diizinkan: %d %q", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
	}
	w = do(h, "OPTIONS", "/api/login", "", map[string]string{"Origin": "http://localhost:5173", "Access-Control-Request-Method": "POST"})
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("origin frontend ditolak: %d", w.Code)
	}
}

func TestLoginHarusJSONDanOriginSah(t *testing.T) {
	h := setup()
	if w := do(h, "POST", "/api/login", "username=a&password=b", map[string]string{"Content-Type": "application/x-www-form-urlencoded"}); w.Code != 415 {
		t.Errorf("login form-urlencoded = %d, mau 415", w.Code)
	}
	if w := do(h, "POST", "/api/login", `{}`, map[string]string{"Content-Type": "application/json", "Origin": "https://jahat.example"}); w.Code != 403 {
		t.Errorf("login dari origin asing = %d, mau 403", w.Code)
	}
	if w := do(h, "POST", "/api/login", `{"username":"","password":""}`, map[string]string{"Content-Type": "application/json"}); w.Code != 401 {
		t.Errorf("login kosong = %d, mau 401", w.Code)
	}
}

func TestBodyTerlaluBesar(t *testing.T) {
	big := `{"username":"` + strings.Repeat("a", 2<<20) + `"}`
	if w := do(setup(), "POST", "/api/login", big, map[string]string{"Content-Type": "application/json"}); w.Code != 400 {
		t.Errorf("body 2MB = %d, mau 400", w.Code)
	}
}
