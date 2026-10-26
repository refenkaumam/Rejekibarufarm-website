package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

func arahParam(c *gin.Context) (string, bool) {
	arah := c.DefaultQuery("arah", "masuk")
	if arah != "masuk" && arah != "keluar" {
		bad(c, "Arah tidak valid")
		return "", false
	}
	return arah, true
}

// ---------- Susu Melk ----------
func ListMelkSusu(c *gin.Context) {
	arah, valid := arahParam(c)
	if !valid {
		return
	}
	rows, err := database.DB.Query(`SELECT id,tgl,liter,harga,IFNULL(pembeli,'') FROM melk_susu
		WHERE arah=? ORDER BY tgl DESC,id DESC LIMIT 100`, arah)
	if err != nil {
		fail(c, "list susu melk", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, pembeli string
		var l, h float64
		if err := rows.Scan(&id, &tgl, &l, &h, &pembeli); err != nil {
			fail(c, "list susu melk", err)
			return
		}
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
	if !bindJSON(c, &in) {
		return
	}
	in.Pembeli = trim(in.Pembeli)
	if in.Tgl == "" || in.Liter <= 0 || in.Harga < 0 || in.Pembeli == "" {
		bad(c, "Tanggal, liter, harga, dan pembeli wajib diisi")
		return
	}
	total := in.Liter * in.Harga
	if !validTgl(in.Tgl) || !textOK(in.Pembeli, 100) ||
		!inRange(in.Liter, 0.01, maxLiter) || !inRange(in.Harga, 0, maxUang) || !inRange(total, 0, maxUang) {
		bad(c, "Tanggal, liter, harga, atau nama pembeli tidak valid")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO melk_susu (tgl,arah,liter,harga,pembeli) VALUES (?,'keluar',?,?,?)", in.Tgl, in.Liter, total, in.Pembeli)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'melk','masuk','jual_susu',?,?,'melk_susu',?)`, in.Tgl, total, "Jual susu ke "+in.Pembeli, id); err != nil {
			return err
		}
		after, err := services.Snapshot(tx, "melk_susu", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "melk_susu", id, nil, after)
	})
	if done(c, "tambah susu melk", err) {
		ok(c)
	}
}

func DelMelkSusu(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		rows, err := services.SnapshotWhere(tx, "melk_susu", "id=? AND arah='keluar'", id)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return errBad("Susu masuk dihapus dari menu Penjualan Susu di kandang")
		}
		kas, err := services.SnapshotWhere(tx, "kas", "ref_tabel='melk_susu' AND ref_id=?", id)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM melk_susu WHERE id=? AND arah='keluar'", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM kas WHERE ref_tabel='melk_susu' AND ref_id=?", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "melk_susu", id, map[string]any{"melk_susu": rows[0], "kas": kas}, nil)
	})
	if done(c, "hapus susu melk", err) {
		ok(c)
	}
}
