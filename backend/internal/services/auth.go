package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rejeki-baru-farm/internal/database"
)

const (
	BcryptCost = 12
	CookieName = "rbf_session"
	issuer     = "rejeki-baru-farm"
)

var (
	ErrBadCredentials = errors.New("username atau password salah")
	ErrLocked         = errors.New("terlalu banyak percobaan login")
	ErrNoSession      = errors.New("sesi tidak valid")
	ErrWeakPassword   = errors.New("Password minimal 10 dan maksimal 72 karakter")
)

type AuthOptions struct {
	Secret         string
	TTL            time.Duration
	CookieSecure   bool
	CookieSameSite string
}

var (
	secret         []byte
	ttl            time.Duration
	cookieSecure   bool
	cookieSameSite http.SameSite
	dummyHash      []byte
)

func InitAuth(o AuthOptions) {
	secret = []byte(o.Secret)
	ttl = o.TTL
	cookieSecure = o.CookieSecure
	switch o.CookieSameSite {
	case "strict":
		cookieSameSite = http.SameSiteStrictMode
	case "none":
		cookieSameSite = http.SameSiteNoneMode
	default:
		cookieSameSite = http.SameSiteLaxMode
	}
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("pembanding-waktu-login"), BcryptCost)
}

func RandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Session adalah identitas user yang sedang login, selalu dibaca ulang dari database.
type Session struct {
	UserID   int64
	Username string
	Role     string
	Menus    []string
	JTI      string
}

type claims struct {
	Ver int `json:"ver"`
	jwt.RegisteredClaims
}

func issueToken(uid int64, ver int) (token, jti string, exp time.Time, err error) {
	now := time.Now()
	jti = RandomHex(16)
	exp = now.Add(ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{Ver: ver, RegisteredClaims: jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   strconv.FormatInt(uid, 10),
		ID:        jti,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}})
	token, err = t.SignedString(secret)
	return
}

func parseToken(s string) (*claims, bool) {
	var cl claims
	t, err := jwt.ParseWithClaims(s, &cl, func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !t.Valid || cl.ID == "" {
		return nil, false
	}
	return &cl, true
}

// LoadSession memvalidasi token lalu mengambil role & menu terbaru dari database.
// Token dari user yang sudah dihapus, logout, atau ganti password otomatis ditolak.
func LoadSession(token string) (Session, error) {
	if token == "" {
		return Session{}, ErrNoSession
	}
	cl, ok := parseToken(token)
	if !ok {
		return Session{}, ErrNoSession
	}
	uid, err := strconv.ParseInt(cl.Subject, 10, 64)
	if err != nil {
		return Session{}, ErrNoSession
	}
	var s Session
	var m sql.NullString
	var ver int
	err = database.DB.QueryRow("SELECT id,username,role,menus,token_version FROM users WHERE id=?", uid).
		Scan(&s.UserID, &s.Username, &s.Role, &m, &ver)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNoSession
	}
	if err != nil {
		return Session{}, err
	}
	if ver != cl.Ver {
		return Session{}, ErrNoSession
	}
	s.Menus = Efektif(s.Role, m)
	s.JTI = cl.ID
	return s, nil
}

// CSRFToken diturunkan dari ID sesi, sehingga tidak perlu disimpan di server.
func CSRFToken(jti string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("csrf:" + jti))
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidCSRF(jti, got string) bool {
	return got != "" && hmac.Equal([]byte(CSRFToken(jti)), []byte(got))
}

func SetSessionCookie(w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: token, Path: "/", Expires: exp,
		MaxAge: int(time.Until(exp).Seconds()), HttpOnly: true, Secure: cookieSecure, SameSite: cookieSameSite})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: cookieSecure, SameSite: cookieSameSite})
}

type LoginResult struct {
	Session Session
	Token   string
	Expires time.Time
}

// Login memeriksa password (waktu respons dibuat sama untuk user yang tidak ada) dan membuat sesi baru.
func Login(username, password, ip string) (LoginResult, error) {
	if !LoginAllowed(ip, username) {
		return LoginResult{}, ErrLocked
	}
	var s Session
	var hash string
	var m sql.NullString
	var ver int
	err := database.DB.QueryRow("SELECT id,username,password_hash,role,menus,token_version FROM users WHERE username=?", username).
		Scan(&s.UserID, &s.Username, &hash, &s.Role, &m, &ver)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LoginResult{}, err
	}
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		LoginFailed(ip, username)
		return LoginResult{}, ErrBadCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		LoginFailed(ip, username)
		return LoginResult{}, ErrBadCredentials
	}
	LoginSucceeded(ip, username)

	// Hash lama dengan cost rendah dinaikkan saat user berhasil login
	if c, err := bcrypt.Cost([]byte(hash)); err == nil && c < BcryptCost {
		if nh, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost); err == nil {
			if _, err := database.DB.Exec("UPDATE users SET password_hash=? WHERE id=?", string(nh), s.UserID); err != nil {
				log.Printf("login: gagal memperbarui hash password user %d: %v", s.UserID, err)
			}
		}
	}
	if s.Username == "admin" && password == "admin123" {
		log.Println("PERINGATAN: akun 'admin' masih memakai password bawaan. Segera ganti password-nya.")
	}

	token, jti, exp, err := issueToken(s.UserID, ver)
	if err != nil {
		return LoginResult{}, err
	}
	s.Menus = Efektif(s.Role, m)
	s.JTI = jti
	return LoginResult{Session: s, Token: token, Expires: exp}, nil
}

// Logout membatalkan semua sesi milik user (token_version dinaikkan).
func Logout(uid int64) error {
	_, err := database.DB.Exec("UPDATE users SET token_version=token_version+1 WHERE id=?", uid)
	return err
}

func CheckPassword(p string) error {
	if len(p) < 10 || len(p) > 72 {
		return ErrWeakPassword
	}
	return nil
}

func HashPassword(p string) (string, error) {
	if err := CheckPassword(p); err != nil {
		return "", err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(p), BcryptCost)
	return string(h), err
}

// ChangePassword mengganti password milik user sendiri, sesi lain otomatis keluar.
func ChangePassword(uid int64, oldPw, newPw string) (LoginResult, error) {
	h, err := HashPassword(newPw)
	if err != nil {
		return LoginResult{}, err
	}
	var s Session
	var hash string
	var m sql.NullString
	var ver int
	if err := database.DB.QueryRow("SELECT id,username,password_hash,role,menus,token_version FROM users WHERE id=?", uid).
		Scan(&s.UserID, &s.Username, &hash, &s.Role, &m, &ver); err != nil {
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPw)) != nil {
		return LoginResult{}, ErrBadCredentials
	}
	if _, err := database.DB.Exec("UPDATE users SET password_hash=?, token_version=token_version+1 WHERE id=?", h, uid); err != nil {
		return LoginResult{}, err
	}
	token, jti, exp, err := issueToken(uid, ver+1)
	if err != nil {
		return LoginResult{}, err
	}
	s.Menus = Efektif(s.Role, m)
	s.JTI = jti
	return LoginResult{Session: s, Token: token, Expires: exp}, nil
}

// SetPassword dipakai super admin untuk mereset password user lain; sesi user itu ikut berakhir.
func SetPassword(uid int64, newPw string) error {
	h, err := HashPassword(newPw)
	if err != nil {
		return err
	}
	res, err := database.DB.Exec("UPDATE users SET password_hash=?, token_version=token_version+1 WHERE id=?", h, uid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SeedAdmin membuat akun awal kalau tabel users masih kosong.
// Password diambil dari ADMIN_INITIAL_PASSWORD, atau dibuat acak dan ditampilkan sekali di log server.
func SeedAdmin(initial string) error {
	var n int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	pw, generated := initial, false
	if pw == "" {
		pw, generated = RandomHex(8), true
	}
	h, err := HashPassword(pw)
	if err != nil {
		return errors.New("ADMIN_INITIAL_PASSWORD: " + err.Error())
	}
	if _, err := database.DB.Exec("INSERT INTO users (username,password_hash,role) VALUES ('admin',?,'super_admin')", h); err != nil {
		return err
	}
	if generated {
		log.Printf("Akun awal dibuat: username 'admin', password sementara %s (hanya ditampilkan sekali, segera ganti)", pw)
	} else {
		log.Println("Akun awal 'admin' dibuat dengan password dari ADMIN_INITIAL_PASSWORD. Hapus variabel itu setelah login pertama.")
	}
	return nil
}
