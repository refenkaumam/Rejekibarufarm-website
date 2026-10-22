package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

func ListProduksi(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT tgl, SUM(IF(sesi='pagi',liter,0)), SUM(IF(sesi='sore',liter,0))
		FROM produksi_susu GROUP BY tgl ORDER BY tgl DESC LIMIT 31`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var t string
		var pg, sr float64
		rows.Scan(&t, &pg, &sr)
		out = append(out, gin.H{"tgl": t, "pagi": pg, "sore": sr, "total": pg + sr})
	}
	c.JSON(200, out)
}

func SaveProduksi(c *gin.Context) {
	var in struct {
		Tgl   string  `json:"tgl"`
		Sesi  string  `json:"sesi"`
		Liter float64 `json:"liter"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Tgl == "" || (in.Sesi != "pagi" && in.Sesi != "sore") || in.Liter < 0 {
		c.JSON(400, gin.H{"error": "Tanggal, sesi, dan liter wajib diisi"})
		return
	}
	_, err := database.DB.Exec(`INSERT INTO produksi_susu (tgl,sesi,liter) VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE liter=VALUES(liter)`, in.Tgl, in.Sesi, in.Liter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal simpan"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func RekapProduksi(c *gin.Context) {
	f := "%Y-%m"
	if c.Query("by") == "tahun" {
		f = "%Y"
	}
	rows, err := database.DB.Query(`SELECT DATE_FORMAT(tgl,?) p, SUM(liter), COUNT(DISTINCT tgl)
		FROM produksi_susu GROUP BY p ORDER BY p DESC LIMIT 24`, f)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var p string
		var t float64
		var h int
		rows.Scan(&p, &t, &h)
		out = append(out, gin.H{"periode": p, "total": t, "hari": h})
	}
	c.JSON(200, out)
}
