package middleware

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/services"
)

func mutating(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
}

// Auth memastikan request membawa cookie sesi yang valid. Identitas, role, dan menu
// selalu dibaca dari database. Request yang mengubah data wajib membawa header X-CSRF-Token.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(services.CookieName)
		s, err := services.LoadSession(token)
		if err != nil {
			if errors.Is(err, services.ErrNoSession) {
				if token != "" {
					services.ClearSessionCookie(c.Writer)
				}
				c.AbortWithStatusJSON(401, gin.H{"error": "Belum login"})
				return
			}
			log.Printf("auth: %v", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "Terjadi kesalahan server"})
			return
		}
		if mutating(c.Request.Method) && !services.ValidCSRF(s.JTI, c.GetHeader("X-CSRF-Token")) {
			c.AbortWithStatusJSON(403, gin.H{"error": "Sesi tidak valid, muat ulang halaman lalu coba lagi"})
			return
		}
		c.Set("uid", s.UserID)
		c.Set("username", s.Username)
		c.Set("role", s.Role)
		c.Set("menus", s.Menus)
		c.Set("jti", s.JTI)
		c.Next()
	}
}

// Allow membatasi berdasarkan role (role dari database, bukan dari token).
func Allow(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, r := range roles {
			if c.GetString("role") == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "Tidak punya akses"})
	}
}

// Akses membatasi berdasarkan menu yang dicentang untuk akun tersebut.
func Akses(keys ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		menus := c.GetStringSlice("menus")
		for _, k := range keys {
			for _, m := range menus {
				if m == k {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "Tidak punya akses"})
	}
}
