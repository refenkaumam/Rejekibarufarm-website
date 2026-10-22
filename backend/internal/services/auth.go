package services

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rejeki-baru-farm/internal/database"
)

var secret []byte

func InitAuth(s string) { secret = []byte(s) }

func IssueToken(username, role string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"u": username, "r": role, "exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return t.SignedString(secret)
}

func ParseToken(s string) (username, role string, ok bool) {
	t, err := jwt.Parse(s, func(*jwt.Token) (any, error) { return secret, nil }, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !t.Valid {
		return "", "", false
	}
	m, isMap := t.Claims.(jwt.MapClaims)
	if !isMap {
		return "", "", false
	}
	username, _ = m["u"].(string)
	role, _ = m["r"].(string)
	return username, role, username != ""
}

// Login memeriksa password dan mengembalikan token, role, dan daftar menu.
func Login(username, password string) (string, string, []string, bool) {
	var hash, role string
	err := database.DB.QueryRow("SELECT password_hash, role FROM users WHERE username=?", username).Scan(&hash, &role)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", "", nil, false
	}
	token, err := IssueToken(username, role)
	if err != nil {
		return "", "", nil, false
	}
	_, menus, _ := MenusOf(username)
	return token, role, menus, true
}

// SeedAdmin membuat akun awal kalau tabel users masih kosong.
func SeedAdmin() {
	var n int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if n > 0 {
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	database.DB.Exec("INSERT INTO users (username,password_hash,role) VALUES ('admin',?,'super_admin')", string(h))
	log.Println("User awal dibuat: admin / admin123 (segera ganti)")
}
