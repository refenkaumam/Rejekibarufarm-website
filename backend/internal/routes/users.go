package routes

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
)

func registerUsers(api *gin.RouterGroup) {
	u := api.Group("/users", middleware.Allow("super_admin"))
	u.GET("", controllers.ListUsers)
	u.POST("", controllers.AddUser)
	u.PUT("/:id/menus", controllers.SetMenus)
	u.DELETE("/:id", controllers.DelUser)
}
