package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// ---------- Akun (super admin) ----------
func ListUsers(c *gin.Context) {
	rows, _ := database.DB.Query("SELECT id,username,role,menus FROM users ORDER BY username")
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var u, r string
		var m sql.NullString
		rows.Scan(&id, &u, &r, &m)
		out = append(out, gin.H{"id": id, "username": u, "role": r, "menus": services.Efektif(r, m)})
	}
	c.JSON(200, out)
}

func AddUser(c *gin.Context) {
	var in struct {
		Username, Password, Role string
		Menus                    []string
	}
	if c.BindJSON(&in) != nil {
		return
	}
	valid := map[string]bool{"super_admin": true, "owner": true, "peternakan": true, "melk": true, "back_office": true}
	if in.Username == "" || len(in.Password) < 6 || !valid[in.Role] {
		c.JSON(400, gin.H{"error": "Username, role, dan password (minimal 6 karakter) wajib diisi"})
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	m := sql.NullString{String: services.CleanMenus(in.Menus), Valid: in.Menus != nil}
	if _, err := database.DB.Exec("INSERT INTO users (username,password_hash,role,menus) VALUES (?,?,?,?)", in.Username, string(h), in.Role, m); err != nil {
		c.JSON(400, gin.H{"error": "Username sudah dipakai"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func SetMenus(c *gin.Context) {
	var in struct{ Menus []string }
	if c.BindJSON(&in) != nil {
		return
	}
	database.DB.Exec("UPDATE users SET menus=? WHERE id=? AND role<>'super_admin'", services.CleanMenus(in.Menus), c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}

func DelUser(c *gin.Context) {
	var name string
	database.DB.QueryRow("SELECT username FROM users WHERE id=?", c.Param("id")).Scan(&name)
	if name == c.GetString("username") {
		c.JSON(400, gin.H{"error": "Tidak bisa menghapus akun yang sedang dipakai"})
		return
	}
	database.DB.Exec("DELETE FROM users WHERE id=?", c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}
