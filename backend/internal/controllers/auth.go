package controllers

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

const pesanLoginGagal = "Username atau password salah"

func auditLogin(c *gin.Context, a services.Actor, aksi string) {
	if err := services.Audit(database.DB, a, aksi, "users", a.UserID, nil, nil); err != nil {
		log.Printf("audit %s: %v", aksi, err)
	}
}

func Login(c *gin.Context) {
	var in struct{ Username, Password string }
	if !bindJSON(c, &in) {
		return
	}
	username := trim(in.Username)
	attempt := services.Actor{Username: username, IP: c.ClientIP()}
	if username == "" || in.Password == "" || len(username) > 50 || len(in.Password) > 200 {
		c.JSON(401, gin.H{"error": pesanLoginGagal})
		return
	}
	res, err := services.Login(username, in.Password, c.ClientIP())
	switch {
	case errors.Is(err, services.ErrLocked):
		auditLogin(c, attempt, "login_diblokir")
		c.JSON(429, gin.H{"error": "Terlalu banyak percobaan login. Coba lagi dalam 15 menit."})
		return
	case errors.Is(err, services.ErrBadCredentials):
		auditLogin(c, attempt, "login_gagal")
		c.JSON(401, gin.H{"error": pesanLoginGagal})
		return
	case err != nil:
		fail(c, "login", err)
		return
	}
	s := res.Session
	services.SetSessionCookie(c.Writer, res.Token, res.Expires)
	auditLogin(c, services.Actor{UserID: s.UserID, Username: s.Username, IP: c.ClientIP()}, "login_berhasil")
	c.JSON(200, gin.H{"username": s.Username, "role": s.Role, "menus": s.Menus, "csrf": services.CSRFToken(s.JTI)})
}

func Me(c *gin.Context) {
	c.JSON(200, gin.H{"username": c.GetString("username"), "role": c.GetString("role"),
		"menus": c.GetStringSlice("menus"), "csrf": services.CSRFToken(c.GetString("jti"))})
}

func Logout(c *gin.Context) {
	if err := services.Logout(c.GetInt64("uid")); err != nil {
		fail(c, "logout", err)
		return
	}
	services.ClearSessionCookie(c.Writer)
	auditLogin(c, actor(c), "logout")
	ok(c)
}

// ChangePassword: user mengganti password sendiri. Sesi di perangkat lain otomatis keluar.
func ChangePassword(c *gin.Context) {
	var in struct {
		Lama string `json:"password_lama"`
		Baru string `json:"password_baru"`
	}
	if !bindJSON(c, &in) {
		return
	}
	if err := services.CheckPassword(in.Baru); err != nil {
		bad(c, err.Error())
		return
	}
	res, err := services.ChangePassword(c.GetInt64("uid"), in.Lama, in.Baru)
	if errors.Is(err, services.ErrBadCredentials) {
		bad(c, "Password lama salah")
		return
	}
	if err != nil {
		fail(c, "ganti password", err)
		return
	}
	services.SetSessionCookie(c.Writer, res.Token, res.Expires)
	auditLogin(c, actor(c), "ganti_password")
	c.JSON(200, gin.H{"ok": true, "csrf": services.CSRFToken(res.Session.JTI)})
}
