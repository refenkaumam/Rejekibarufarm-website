package controllers

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// Batas atas nilai yang masuk akal (kolom DECIMAL(15,2) menampung < 1e13)
const (
	maxUang  = 1e12
	maxLiter = 1e6
	maxKg    = 1e7
	maxGram  = 1e9
)

var errNotFound = errors.New("data tidak ditemukan")

// errBad adalah kesalahan input yang pesannya aman ditampilkan ke user.
type errBad string

func (e errBad) Error() string { return string(e) }

func bindJSON(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		bad(c, "Format data tidak valid")
		return false
	}
	return true
}

func bad(c *gin.Context, msg string) { c.JSON(400, gin.H{"error": msg}) }

func ok(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) }

// fail mencatat detail error di log server dan mengirim pesan umum ke client.
func fail(c *gin.Context, where string, err error) {
	log.Printf("%s: %v", where, err)
	c.JSON(500, gin.H{"error": "Terjadi kesalahan server"})
}

// done menerjemahkan hasil transaksi menjadi respons; true kalau berhasil.
func done(c *gin.Context, where string, err error) bool {
	var b errBad
	switch {
	case err == nil:
		return true
	case errors.Is(err, errNotFound):
		c.JSON(404, gin.H{"error": "Data tidak ditemukan"})
	case errors.As(err, &b):
		bad(c, string(b))
	default:
		fail(c, where, err)
	}
	return false
}

func paramID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		bad(c, "ID tidak valid")
		return 0, false
	}
	return id, true
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func withTx(fn func(tx *sql.Tx) error) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func actor(c *gin.Context) services.Actor {
	return services.Actor{UserID: c.GetInt64("uid"), Username: c.GetString("username"), IP: c.ClientIP()}
}

func audit(ex services.Execer, c *gin.Context, aksi, tabel string, id int64, before, after any) error {
	return services.Audit(ex, actor(c), aksi, tabel, id, before, after)
}

// ---------- validasi ----------

func today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func parseTgl(s string) (time.Time, bool) {
	if len(s) != 10 {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	return t, err == nil
}

// validTgl: tanggal transaksi antara 1 Jan 2000 dan 1 tahun ke depan
func validTgl(s string) bool {
	t, ok := parseTgl(s)
	return ok && t.Year() >= 2000 && !t.After(today().AddDate(1, 0, 0))
}

func validBulan(s string) bool {
	if len(s) != 7 {
		return false
	}
	t, err := time.Parse("2006-01", s)
	return err == nil && t.Year() >= 2000 && t.Year() <= 2100
}

func bulanParam(c *gin.Context) (string, bool) {
	b := c.Query("bulan")
	if b == "" {
		b = time.Now().Format("2006-01")
	}
	if !validBulan(b) {
		bad(c, "Format bulan tidak valid")
		return "", false
	}
	return b, true
}

func inRange(v, min, max float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= min && v <= max
}

func textOK(s string, max int) bool {
	return utf8.ValidString(s) && utf8.RuneCountInString(s) <= max
}

func trim(s string) string { return strings.TrimSpace(s) }
