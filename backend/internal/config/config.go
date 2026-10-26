package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env            string
	DSN            string
	JWTSecret      string
	Host           string
	Port           string
	AllowedOrigins []string
	TrustedProxies []string
	CookieSecure   bool
	CookieSameSite string
	SessionTTL     time.Duration
	AdminPassword  string
}

func (c Config) Production() bool { return c.Env == "production" }

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func list(s string) []string {
	out := []string{}
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, strings.TrimRight(v, "/"))
		}
	}
	return out
}

// loadDotEnv membaca file .env (jika ada) tanpa menimpa environment variable yang sudah diset.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(strings.TrimPrefix(k, "export "))
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, set := os.LookupEnv(k); !set {
			os.Setenv(k, v)
		}
	}
}

// Load membaca konfigurasi dari environment variable (dan file .env bila ada).
// Di produksi (APP_ENV=production) konfigurasi penting wajib diisi, tidak ada nilai bawaan.
func Load() Config {
	loadDotEnv(".env")
	c := Config{
		Env:            strings.ToLower(get("APP_ENV", "development")),
		DSN:            os.Getenv("DB_DSN"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		Host:           os.Getenv("HOST"),
		Port:           get("PORT", "8080"),
		AllowedOrigins: list(os.Getenv("ALLOWED_ORIGINS")),
		TrustedProxies: list(os.Getenv("TRUSTED_PROXIES")),
		CookieSameSite: strings.ToLower(get("COOKIE_SAMESITE", "lax")),
		AdminPassword:  os.Getenv("ADMIN_INITIAL_PASSWORD"),
		SessionTTL:     12 * time.Hour,
	}
	if v := os.Getenv("SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < 15*time.Minute || d > 7*24*time.Hour {
			log.Fatal("SESSION_TTL tidak valid (contoh: 12h, antara 15m dan 168h)")
		}
		c.SessionTTL = d
	}
	defSecure := "false"
	if c.Production() {
		defSecure = "true"
	}
	c.CookieSecure = strings.ToLower(get("COOKIE_SECURE", defSecure)) == "true"

	weakSecret := len(c.JWTSecret) < 32 || c.JWTSecret == "ganti-secret-ini"
	if c.Production() {
		missing := []string{}
		if c.DSN == "" {
			missing = append(missing, "DB_DSN")
		}
		if weakSecret {
			missing = append(missing, "JWT_SECRET (minimal 32 karakter acak)")
		}
		if len(c.AllowedOrigins) == 0 {
			missing = append(missing, "ALLOWED_ORIGINS")
		}
		if len(missing) > 0 {
			log.Fatalf("Konfigurasi produksi belum lengkap: %s", strings.Join(missing, ", "))
		}
		if !c.CookieSecure {
			log.Println("PERINGATAN: COOKIE_SECURE=false di produksi. Cookie sesi bisa terkirim lewat HTTP biasa.")
		}
	} else {
		if c.DSN == "" {
			c.DSN = "root@tcp(127.0.0.1:3306)/rejeki" // MySQL lokal (mis. Laragon) tanpa password
		}
		if len(c.AllowedOrigins) == 0 {
			c.AllowedOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
		}
		if weakSecret {
			b := make([]byte, 32)
			rand.Read(b)
			c.JWTSecret = hex.EncodeToString(b)
			log.Println("PERINGATAN: JWT_SECRET kosong/lemah, memakai secret acak sementara (semua sesi hilang saat server restart).")
		}
	}

	switch c.CookieSameSite {
	case "lax", "strict":
	case "none":
		if !c.CookieSecure {
			log.Fatal("COOKIE_SAMESITE=none wajib disertai COOKIE_SECURE=true")
		}
	default:
		log.Fatal("COOKIE_SAMESITE harus lax, strict, atau none")
	}
	return c
}
