package config

import "os"

type Config struct {
	DSN       string
	JWTSecret string
	Port      string
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// Load membaca konfigurasi dari environment variable.
func Load() Config {
	return Config{
		DSN:       get("DB_DSN", "root:root@tcp(127.0.0.1:3306)/rejeki"),
		JWTSecret: get("JWT_SECRET", "ganti-secret-ini"),
		Port:      get("PORT", "8080"),
	}
}
