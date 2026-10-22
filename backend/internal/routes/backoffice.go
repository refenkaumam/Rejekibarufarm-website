package routes

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
)

func registerBackOffice(api *gin.RouterGroup) {
	b := api.Group("/bo")
	b.GET("/ringkasan", middleware.Akses("bo"), controllers.BoRingkasan)
	b.GET("/labarugi", middleware.Akses("bo"), controllers.BoLabaRugi)
	b.GET("/kasawal", middleware.Akses("bokas"), controllers.BoKasAwal)
	b.POST("/kasawal", middleware.Akses("bokas"), controllers.BoSetKasAwal)
	b.GET("/kas", middleware.Akses("bokas"), controllers.BoListKas)
	b.POST("/kas", middleware.Akses("bokas"), controllers.BoAddKas)
	b.DELETE("/kas/:id", middleware.Akses("bokas"), controllers.BoDelKas)

	e := api.Group("/bo", middleware.Allow("super_admin", "owner"))
	e.GET("/export", middleware.Akses("bo"), controllers.BoExport)
}
