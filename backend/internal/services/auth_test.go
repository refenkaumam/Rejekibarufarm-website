package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	InitAuth(AuthOptions{Secret: strings.Repeat("k", 40), TTL: time.Hour})
}

func TestTokenRoundTrip(t *testing.T) {
	tok, jti, _, err := issueToken(7, 3)
	if err != nil {
		t.Fatal(err)
	}
	cl, ok := parseToken(tok)
	if !ok || cl.Subject != "7" || cl.Ver != 3 || cl.ID != jti {
		t.Fatalf("token tidak terbaca dengan benar: %+v", cl)
	}
}

func TestTokenDitolak(t *testing.T) {
	// algoritma none
	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "1", "iss": issuer, "jti": "x", "exp": time.Now().Add(time.Hour).Unix()})
	s, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, ok := parseToken(s); ok {
		t.Fatal("token alg=none diterima")
	}
	// secret lain
	other := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "iss": issuer, "jti": "x", "exp": time.Now().Add(time.Hour).Unix()})
	s, _ = other.SignedString([]byte("secret-lain-yang-tidak-dikenal-server"))
	if _, ok := parseToken(s); ok {
		t.Fatal("token dengan secret lain diterima")
	}
	// kedaluwarsa
	old := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "iss": issuer, "jti": "x", "exp": time.Now().Add(-time.Minute).Unix()})
	s, _ = old.SignedString(secret)
	if _, ok := parseToken(s); ok {
		t.Fatal("token kedaluwarsa diterima")
	}
}

func TestCSRF(t *testing.T) {
	tok := CSRFToken("abc")
	if !ValidCSRF("abc", tok) || ValidCSRF("abd", tok) || ValidCSRF("abc", "") {
		t.Fatal("validasi CSRF salah")
	}
}

func TestLoginLockout(t *testing.T) {
	ip, u := "10.0.0.1", "budi"
	for i := 0; i < userFailLimit; i++ {
		if !LoginAllowed(ip, u) {
			t.Fatalf("terblokir terlalu cepat pada percobaan ke-%d", i+1)
		}
		LoginFailed(ip, u)
	}
	if LoginAllowed(ip, u) {
		t.Fatal("tidak terblokir setelah batas gagal")
	}
	if !LoginAllowed("10.0.0.2", u) {
		t.Fatal("IP lain ikut terblokir untuk username yang sama")
	}
}

func TestCheckPassword(t *testing.T) {
	if CheckPassword("pendek") == nil || CheckPassword(strings.Repeat("a", 73)) == nil || CheckPassword("cukup-panjang-123") != nil {
		t.Fatal("aturan panjang password salah")
	}
}
