package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Keju Melk ----------
func ListMelkKeju(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,jumlah_gram,harga,IFNULL(pembeli,'') FROM melk_keju
		WHERE arah=? ORDER BY tgl DESC,id DESC LIMIT 100`, c.DefaultQuery("arah", "masuk"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, pembeli string
		var g, h float64
		rows.Scan(&id, &tgl, &g, &h, &pembeli)
		out = append(out, gin.H{"id": id, "tgl": tgl, "gram": g, "harga": h, "pembeli": pembeli})
	}
	c.JSON(200, out)
}

// Keju masuk (dari kandang) -> kas keluar Melk. Keju keluar -> kas masuk Melk.
func AddKeju(c *gin.Context) {
	var in struct {
		Tgl     string  `json:"tgl"`
		Arah    string  `json:"arah"`
		Gram    float64 `json:"gram"`
		Harga   float64 `json:"harga"`
		Pembeli string  `json:"pembeli"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Tgl == "" || (in.Arah != "masuk" && in.Arah != "keluar") || in.Gram <= 0 || in.Harga < 0 {
		c.JSON(400, gin.H{"error": "Tanggal, arah, jumlah, dan harga wajib diisi"})
		return
	}
	if in.Arah == "keluar" && in.Pembeli == "" {
		c.JSON(400, gin.H{"error": "Pembeli wajib diisi untuk keju keluar"})
		return
	}
	tx, _ := database.DB.Begin()
	res, err := tx.Exec("INSERT INTO melk_keju (tgl,arah,jumlah_gram,harga,pembeli) VALUES (?,?,?,?,NULLIF(?,''))",
		in.Tgl, in.Arah, in.Gram, in.Harga, in.Pembeli)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal simpan"})
		return
	}
	id, _ := res.LastInsertId()
	if in.Harga > 0 {
		arah, kat, ket := "keluar", "beli_keju", "Keju masuk dari kandang"
		if in.Arah == "keluar" {
			arah, kat, ket = "masuk", "jual_keju", "Jual keju ke "+in.Pembeli
		}
		tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'melk',?,?,?,?,'melk_keju',?)`, in.Tgl, arah, kat, in.Harga, ket, id)
	}
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func DelKeju(c *gin.Context) {
	tx, _ := database.DB.Begin()
	tx.Exec("DELETE FROM kas WHERE ref_tabel='melk_keju' AND ref_id=?", c.Param("id"))
	tx.Exec("DELETE FROM melk_keju WHERE id=?", c.Param("id"))
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}
