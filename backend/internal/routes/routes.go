package routes

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS)
	r.POST("/api/login", controllers.Login)

	api := r.Group("/api", middleware.Auth())
	api.GET("/me", controllers.Me)

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
