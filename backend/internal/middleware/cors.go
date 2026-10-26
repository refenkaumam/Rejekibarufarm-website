package middleware

import "github.com/gin-gonic/gin"

// CORS hanya mengizinkan origin frontend yang terdaftar di ALLOWED_ORIGINS (dengan cookie).
func CORS(allowed []string) gin.HandlerFunc {
	ok := map[string]bool{}
	for _, o := range allowed {
		ok[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		h := c.Writer.Header()
		h.Add("Vary", "Origin")
		if origin != "" && ok[origin] {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == "OPTIONS" {
			if origin != "" && !ok[origin] {
				c.AbortWithStatus(403)
				return
			}
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
