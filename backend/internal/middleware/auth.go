package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/services"
)

// Auth memastikan request membawa token login yang valid.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		username, role, ok := services.ParseToken(token)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Belum login"})
			return
		}
		c.Set("username", username)
		c.Set("role", role)
		c.Next()
	}
}

// Allow membatasi berdasarkan role.
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
		if _, menus, ok := services.MenusOf(c.GetString("username")); ok {
			for _, k := range keys {
				for _, m := range menus {
					if m == k {
						c.Next()
						return
					}
				}
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "Tidak punya akses"})
	}
}
