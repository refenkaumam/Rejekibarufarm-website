package routes

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
)

func registerMelk(api *gin.RouterGroup) {
	m := api.Group("/melk")
	m.GET("/susu", middleware.Akses("msusu"), controllers.ListMelkSusu)
	m.POST("/susu", middleware.Akses("msusu"), controllers.AddSusuKeluar)
	m.DELETE("/susu/:id", middleware.Akses("msusu"), controllers.DelMelkSusu)
	m.GET("/keju", middleware.Akses("mkeju"), controllers.ListMelkKeju)
	m.POST("/keju", middleware.Akses("mkeju"), controllers.AddKeju)
	m.DELETE("/keju/:id", middleware.Akses("mkeju"), controllers.DelKeju)
	m.GET("/rekap", middleware.Akses("mrekap"), controllers.RekapMelk)
	m.GET("/kas", middleware.Akses("mkas"), controllers.ListKasMelk)
	m.POST("/kas", middleware.Akses("mkas"), controllers.AddKasMelk)
	m.DELETE("/kas/:id", middleware.Akses("mkas"), controllers.DelKasMelk)
}
