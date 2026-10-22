package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Penjualan susu ----------
func ListPenjualan(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,tujuan,liter,harga_per_liter,liter*harga_per_liter
		FROM penjualan_susu ORDER BY tgl DESC,id DESC LIMIT 100`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, tujuan string
		var l, h, t float64
		rows.Scan(&id, &tgl, &tujuan, &l, &h, &t)
		out = append(out, gin.H{"id": id, "tgl": tgl, "tujuan": tujuan, "liter": l, "harga_per_liter": h, "total": t})
	}
	c.JSON(200, out)
}

// Jual susu -> kas masuk kandang. Ke Melk by RBF -> juga susu masuk + kas keluar Melk.
func AddPenjualan(c *gin.Context) {
	var in struct {
		Tgl    string  `json:"tgl"`
		Tujuan string  `json:"tujuan"`
		Liter  float64 `json:"liter"`
		Harga  float64 `json:"harga_per_liter"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Tgl == "" || (in.Tujuan != "koperasi" && in.Tujuan != "melk_rbf") || in.Liter <= 0 || in.Harga < 0 {
		c.JSON(400, gin.H{"error": "Tanggal, tujuan, liter, dan harga wajib diisi"})
		return
	}
	total := in.Liter * in.Harga
	tx, _ := database.DB.Begin()
	res, err := tx.Exec("INSERT INTO penjualan_susu (tgl,tujuan,liter,harga_per_liter) VALUES (?,?,?,?)", in.Tgl, in.Tujuan, in.Liter, in.Harga)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal simpan"})
		return
	}
	id, _ := res.LastInsertId()
	tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
		VALUES (?,'kandang','masuk','pembelian_susu',?,?,'penjualan_susu',?)`, in.Tgl, total, "Jual susu ke "+in.Tujuan, id)
	if in.Tujuan == "melk_rbf" {
		tx.Exec("INSERT INTO melk_susu (tgl,arah,liter,harga,ref_id) VALUES (?,'masuk',?,?,?)", in.Tgl, in.Liter, total, id)
		tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'melk','keluar','beli_susu',?,'Beli susu dari kandang','penjualan_susu',?)`, in.Tgl, total, id)
	}
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func DelPenjualan(c *gin.Context) {
	tx, _ := database.DB.Begin()
	tx.Exec("DELETE FROM kas WHERE ref_tabel='penjualan_susu' AND ref_id=?", c.Param("id"))
	tx.Exec("DELETE FROM melk_susu WHERE ref_id=?", c.Param("id"))
	tx.Exec("DELETE FROM penjualan_susu WHERE id=?", c.Param("id"))
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}
