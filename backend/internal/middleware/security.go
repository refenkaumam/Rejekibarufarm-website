package middleware

import (
	"log"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/services"
)

// SecurityHeaders menambahkan header keamanan untuk semua respons API.
func SecurityHeaders(hsts bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cache-Control", "no-store")
		if hsts {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// BodyLimit membatasi ukuran body request.
func BodyLimit(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		}
		c.Next()
	}
}

// RequireJSON menolak body selain application/json (mencegah CSRF lewat form/text biasa).
func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if mutating(c.Request.Method) && c.Request.ContentLength != 0 {
			mt, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if err != nil || mt != "application/json" {
				c.AbortWithStatusJSON(415, gin.H{"error": "Format data harus JSON"})
				return
			}
		}
		c.Next()
	}
}

// OriginCheck menolak request pengubah data dari origin yang tidak terdaftar.
func OriginCheck(allowed []string) gin.HandlerFunc {
	ok := map[string]bool{}
	for _, o := range allowed {
		ok[o] = true
	}
	return func(c *gin.Context) {
		if o := c.GetHeader("Origin"); mutating(c.Request.Method) && o != "" && !ok[o] {
			c.AbortWithStatusJSON(403, gin.H{"error": "Origin tidak diizinkan"})
			return
		}
		c.Next()
	}
}

// RateLimit membatasi jumlah request per IP.
func RateLimit(l *services.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(429, gin.H{"error": "Terlalu banyak permintaan, coba lagi sebentar lagi"})
			return
		}
		c.Next()
	}
}

// Recovery menangkap panic tanpa membocorkan detail internal ke client.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err any) {
		log.Printf("panic: %v", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "Terjadi kesalahan server"})
	})
}
