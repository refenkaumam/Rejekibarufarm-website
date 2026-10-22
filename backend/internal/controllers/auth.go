package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/services"
)

func Login(c *gin.Context) {
	var in struct{ Username, Password string }
	if c.BindJSON(&in) != nil {
		return
	}
	token, role, menus, ok := services.Login(in.Username, in.Password)
	if !ok {
		c.JSON(401, gin.H{"error": "Username atau password salah"})
		return
	}
	c.JSON(200, gin.H{"token": token, "username": in.Username, "role": role, "menus": menus})
}

func Me(c *gin.Context) {
	role, menus, _ := services.MenusOf(c.GetString("username"))
	c.JSON(200, gin.H{"username": c.GetString("username"), "role": role, "menus": menus})
}
