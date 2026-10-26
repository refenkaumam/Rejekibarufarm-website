package controllers

import (
	"database/sql"
	"errors"
	"regexp"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9._-]{3,50}$`)

var roleValid = map[string]bool{"super_admin": true, "owner": true, "peternakan": true, "melk": true, "back_office": true}

// ---------- Akun (super admin) ----------
func ListUsers(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id,username,role,menus FROM users ORDER BY username")
	if err != nil {
		fail(c, "list users", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var u, r string
		var m sql.NullString
		if err := rows.Scan(&id, &u, &r, &m); err != nil {
			fail(c, "list users", err)
			return
		}
		out = append(out, gin.H{"id": id, "username": u, "role": r, "menus": services.Efektif(r, m)})
	}
	c.JSON(200, out)
}

func AddUser(c *gin.Context) {
	var in struct {
		Username, Password, Role string
		Menus                    []string
	}
	if !bindJSON(c, &in) {
		return
	}
	in.Username = trim(in.Username)
	if in.Username == "" || in.Password == "" || !roleValid[in.Role] {
		bad(c, "Username, role, dan password wajib diisi")
		return
	}
	if !usernameRe.MatchString(in.Username) {
		bad(c, "Username 3-50 karakter, hanya huruf, angka, titik, garis bawah, atau strip")
		return
	}
	h, err := services.HashPassword(in.Password)
	if err != nil {
		bad(c, err.Error())
		return
	}
	m := sql.NullString{String: services.CleanMenus(in.Menus), Valid: in.Menus != nil}
	err = withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO users (username,password_hash,role,menus) VALUES (?,?,?,?)", in.Username, h, in.Role, m)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		after, err := services.Snapshot(tx, "users", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "users", id, nil, after)
	})
	if isDuplicate(err) {
		bad(c, "Username sudah dipakai")
		return
	}
	if done(c, "tambah user", err) {
		ok(c)
	}
}

func SetMenus(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	var in struct{ Menus []string }
	if !bindJSON(c, &in) {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		before, err := services.Snapshot(tx, "users", id)
		if err != nil {
			return err
		}
		if before == nil {
			return errNotFound
		}
		if _, err := tx.Exec("UPDATE users SET menus=? WHERE id=? AND role<>'super_admin'", services.CleanMenus(in.Menus), id); err != nil {
			return err
		}
		after, err := services.Snapshot(tx, "users", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "ubah_akses", "users", id, before, after)
	})
	if done(c, "atur akses", err) {
		ok(c)
	}
}

// ResetPassword: super admin mengatur password baru untuk user lain.
func ResetPassword(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	var in struct{ Password string }
	if !bindJSON(c, &in) {
		return
	}
	err := services.SetPassword(id, in.Password)
	switch {
	case errors.Is(err, services.ErrWeakPassword):
		bad(c, err.Error())
		return
	case errors.Is(err, sql.ErrNoRows):
		done(c, "", errNotFound)
		return
	case err != nil:
		fail(c, "reset password", err)
		return
	}
	if err := audit(database.DB, c, "reset_password", "users", id, nil, nil); err != nil {
		fail(c, "reset password", err)
		return
	}
	ok(c)
}

func DelUser(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	if id == c.GetInt64("uid") {
		bad(c, "Tidak bisa menghapus akun yang sedang dipakai")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		before, err := services.Snapshot(tx, "users", id)
		if err != nil {
			return err
		}
		if before == nil {
			return errNotFound
		}
		if before["role"] == "super_admin" {
			var n int
			if err := tx.QueryRow("SELECT COUNT(*) FROM users WHERE role='super_admin' FOR UPDATE").Scan(&n); err != nil {
				return err
			}
			if n <= 1 {
				return errBad("Tidak bisa menghapus super admin terakhir")
			}
		}
		if _, err := tx.Exec("DELETE FROM users WHERE id=?", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "users", id, before, nil)
	})
	if done(c, "hapus user", err) {
		ok(c)
	}
}
