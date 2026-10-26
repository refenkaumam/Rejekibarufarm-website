package routes

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/config"
	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
	"rejeki-baru-farm/internal/services"
)

func Setup(cfg config.Config) *gin.Engine {
	if cfg.Production() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// Tanpa TRUSTED_PROXIES, IP client diambil dari koneksi langsung (header X-Forwarded-For diabaikan)
	var proxies []string
	if len(cfg.TrustedProxies) > 0 {
		proxies = cfg.TrustedProxies
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		log.Fatal("TRUSTED_PROXIES tidak valid: ", err)
	}
	r.Use(
		gin.Logger(),
		middleware.Recovery(),
		middleware.SecurityHeaders(cfg.Production() && cfg.CookieSecure),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.RateLimit(services.NewRateLimiter(600, time.Minute)),
		middleware.BodyLimit(1<<20),
		middleware.OriginCheck(cfg.AllowedOrigins),
		middleware.RequireJSON(),
	)
	r.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"error": "Tidak ditemukan"}) })

	r.POST("/api/login", controllers.Login)

	api := r.Group("/api", middleware.Auth())
	api.GET("/me", controllers.Me)
	api.POST("/logout", controllers.Logout)
	api.PUT("/me/password", controllers.ChangePassword)

	sapi := api.Group("/sapi", middleware.Akses("sapi"))
	sapi.GET("", controllers.ListSapi)
	sapi.POST("", controllers.SaveSapi)
	sapi.PUT("/:id", controllers.SaveSapi)
	sapi.DELETE("/:id", controllers.DelSapi)

	registerPakanProduksi(api)
	registerPenjualanInvestorKas(api)
	registerMelk(api)
	registerBackOffice(api)
	registerUsers(api)
	return r
}
