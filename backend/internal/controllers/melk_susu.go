package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Susu Melk ----------
func ListMelkSusu(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,liter,harga,IFNULL(pembeli,'') FROM melk_susu
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
		var l, h float64
		rows.Scan(&id, &tgl, &l, &h, &pembeli)
		out = append(out, gin.H{"id": id, "tgl": tgl, "liter": l, "harga": h, "pembeli": pembeli})
	}
	c.JSON(200, out)
}

// Susu keluar -> kas masuk Melk (jual_susu)
func AddSusuKeluar(c *gin.Context) {
	var in struct {
		Tgl     string  `json:"tgl"`
		Liter   float64 `json:"liter"`
		Harga   float64 `json:"harga_per_liter"`
		Pembeli string  `json:"pembeli"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Tgl == "" || in.Liter <= 0 || in.Harga < 0 || in.Pembeli == "" {
		c.JSON(400, gin.H{"error": "Tanggal, liter, harga, dan pembeli wajib diisi"})
		return
	}
	total := in.Liter * in.Harga
	tx, _ := database.DB.Begin()
	res, err := tx.Exec("INSERT INTO melk_susu (tgl,arah,liter,harga,pembeli) VALUES (?,'keluar',?,?,?)", in.Tgl, in.Liter, total, in.Pembeli)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal simpan"})
		return
	}
	id, _ := res.LastInsertId()
	tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
		VALUES (?,'melk','masuk','jual_susu',?,?,'melk_susu',?)`, in.Tgl, total, "Jual susu ke "+in.Pembeli, id)
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func DelMelkSusu(c *gin.Context) {
	tx, _ := database.DB.Begin()
	res, _ := tx.Exec("DELETE FROM melk_susu WHERE id=? AND arah='keluar'", c.Param("id"))
	if n, _ := res.RowsAffected(); n == 0 {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "Susu masuk dihapus dari menu Penjualan Susu di kandang"})
		return
	}
	tx.Exec("DELETE FROM kas WHERE ref_tabel='melk_susu' AND ref_id=?", c.Param("id"))
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}
