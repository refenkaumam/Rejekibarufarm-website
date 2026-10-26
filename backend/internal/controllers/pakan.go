package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/models"
	"rejeki-baru-farm/internal/services"
)

var jenisPakan = map[string]bool{"rumput": true, "konsentrat": true, "tebon_jagung": true, "mineral": true}

func ListPakan(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,jenis,arah,jumlah,satuan,harga_total,IFNULL(keterangan,'')
		FROM pakan_mutasi ORDER BY tgl DESC,id DESC LIMIT 100`)
	if err != nil {
		fail(c, "list pakan", err)
		return
	}
	defer rows.Close()
	out := []models.Pakan{}
	for rows.Next() {
		var p models.Pakan
		if err := rows.Scan(&p.ID, &p.Tgl, &p.Jenis, &p.Arah, &p.Jumlah, &p.Satuan, &p.HargaTotal, &p.Keterangan); err != nil {
			fail(c, "list pakan", err)
			return
		}
		out = append(out, p)
	}
	c.JSON(200, out)
}

// Pakan masuk dengan harga -> otomatis kas keluar kandang (beli_pakan)
func AddPakan(c *gin.Context) {
	var p struct {
		Tgl        string  `json:"tgl"`
		Jenis      string  `json:"jenis"`
		Arah       string  `json:"arah"`
		Jumlah     float64 `json:"jumlah"`
		Satuan     string  `json:"satuan"`
		HargaTotal float64 `json:"harga_total"`
		Keterangan string  `json:"keterangan"`
	}
	if !bindJSON(c, &p) {
		return
	}
	if p.Tgl == "" || !jenisPakan[p.Jenis] || (p.Arah != "masuk" && p.Arah != "keluar") || p.Jumlah <= 0 {
		bad(c, "Tanggal, jenis, arah, dan jumlah wajib diisi")
		return
	}
	p.Satuan, p.Keterangan = trim(p.Satuan), trim(p.Keterangan)
	if p.Satuan == "" {
		p.Satuan = "kg"
	}
	switch {
	case !validTgl(p.Tgl):
		bad(c, "Tanggal tidak valid")
		return
	case !inRange(p.Jumlah, 0.01, maxKg):
		bad(c, "Jumlah pakan di luar batas wajar")
		return
	case !inRange(p.HargaTotal, 0, maxUang):
		bad(c, "Total harga tidak valid")
		return
	case !textOK(p.Satuan, 10) || !textOK(p.Keterangan, 255):
		bad(c, "Satuan maksimal 10 karakter dan keterangan maksimal 255 karakter")
		return
	}
	if p.Arah == "keluar" {
		p.HargaTotal = 0
	}
	err := withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`INSERT INTO pakan_mutasi (tgl,jenis,arah,jumlah,satuan,harga_total,keterangan) VALUES (?,?,?,?,?,?,?)`,
			p.Tgl, p.Jenis, p.Arah, p.Jumlah, p.Satuan, p.HargaTotal, p.Keterangan)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if p.Arah == "masuk" && p.HargaTotal > 0 {
			if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
				VALUES (?,'kandang','keluar','beli_pakan',?,?,'pakan_mutasi',?)`,
				p.Tgl, p.HargaTotal, "Beli "+p.Jenis, id); err != nil {
				return err
			}
		}
		after, err := services.Snapshot(tx, "pakan_mutasi", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "pakan_mutasi", id, nil, after)
	})
	if done(c, "tambah pakan", err) {
		ok(c)
	}
}

func DelPakan(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		row, err := services.Snapshot(tx, "pakan_mutasi", id)
		if err != nil {
			return err
		}
		if row == nil {
			return errNotFound
		}
		kas, err := services.SnapshotWhere(tx, "kas", "ref_tabel='pakan_mutasi' AND ref_id=?", id)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM kas WHERE ref_tabel='pakan_mutasi' AND ref_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM pakan_mutasi WHERE id=?", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "pakan_mutasi", id, map[string]any{"pakan_mutasi": row, "kas": kas}, nil)
	})
	if done(c, "hapus pakan", err) {
		ok(c)
	}
}

func StokPakan(c *gin.Context) {
	out := gin.H{"rumput": 0.0, "konsentrat": 0.0, "tebon_jagung": 0.0, "mineral": 0.0}
	rows, err := database.DB.Query(`SELECT jenis, SUM(IF(arah='masuk',jumlah,-jumlah)) FROM pakan_mutasi GROUP BY jenis`)
	if err != nil {
		fail(c, "stok pakan", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var j string
		var s float64
		if err := rows.Scan(&j, &s); err != nil {
			fail(c, "stok pakan", err)
			return
		}
		out[j] = s
	}
	c.JSON(200, out)
}
