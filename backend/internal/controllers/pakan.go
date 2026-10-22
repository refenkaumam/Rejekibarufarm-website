package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/models"
)

func ListPakan(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,jenis,arah,jumlah,satuan,harga_total,IFNULL(keterangan,'')
		FROM pakan_mutasi ORDER BY tgl DESC,id DESC LIMIT 100`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []models.Pakan{}
	for rows.Next() {
		var p models.Pakan
		rows.Scan(&p.ID, &p.Tgl, &p.Jenis, &p.Arah, &p.Jumlah, &p.Satuan, &p.HargaTotal, &p.Keterangan)
		out = append(out, p)
	}
	c.JSON(200, out)
}

// Pakan masuk dengan harga -> otomatis kas keluar kandang (beli_pakan)
func AddPakan(c *gin.Context) {
	var p models.Pakan
	if c.BindJSON(&p) != nil {
		return
	}
	jenis := map[string]bool{"rumput": true, "konsentrat": true, "tebon_jagung": true, "mineral": true}
	if p.Tgl == "" || !jenis[p.Jenis] || (p.Arah != "masuk" && p.Arah != "keluar") || p.Jumlah <= 0 {
		c.JSON(400, gin.H{"error": "Tanggal, jenis, arah, dan jumlah wajib diisi"})
		return
	}
	if p.Satuan == "" {
		p.Satuan = "kg"
	}
	if p.Arah == "keluar" {
		p.HargaTotal = 0
	}
	tx, _ := database.DB.Begin()
	res, err := tx.Exec(`INSERT INTO pakan_mutasi (tgl,jenis,arah,jumlah,satuan,harga_total,keterangan) VALUES (?,?,?,?,?,?,?)`,
		p.Tgl, p.Jenis, p.Arah, p.Jumlah, p.Satuan, p.HargaTotal, p.Keterangan)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal simpan"})
		return
	}
	if p.Arah == "masuk" && p.HargaTotal > 0 {
		id, _ := res.LastInsertId()
		tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'kandang','keluar','beli_pakan',?,?,'pakan_mutasi',?)`,
			p.Tgl, p.HargaTotal, "Beli "+p.Jenis, id)
	}
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func DelPakan(c *gin.Context) {
	tx, _ := database.DB.Begin()
	tx.Exec("DELETE FROM kas WHERE ref_tabel='pakan_mutasi' AND ref_id=?", c.Param("id"))
	tx.Exec("DELETE FROM pakan_mutasi WHERE id=?", c.Param("id"))
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func StokPakan(c *gin.Context) {
	out := gin.H{"rumput": 0.0, "konsentrat": 0.0, "tebon_jagung": 0.0, "mineral": 0.0}
	rows, err := database.DB.Query(`SELECT jenis, SUM(IF(arah='masuk',jumlah,-jumlah)) FROM pakan_mutasi GROUP BY jenis`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var j string
			var s float64
			rows.Scan(&j, &s)
			out[j] = s
		}
	}
	c.JSON(200, out)
}
