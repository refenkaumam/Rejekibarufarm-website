package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// ---------- Keju Melk ----------
func ListMelkKeju(c *gin.Context) {
	arah, valid := arahParam(c)
	if !valid {
		return
	}
	rows, err := database.DB.Query(`SELECT id,tgl,jumlah_gram,harga,IFNULL(pembeli,'') FROM melk_keju
		WHERE arah=? ORDER BY tgl DESC,id DESC LIMIT 100`, arah)
	if err != nil {
		fail(c, "list keju", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, pembeli string
		var g, h float64
		if err := rows.Scan(&id, &tgl, &g, &h, &pembeli); err != nil {
			fail(c, "list keju", err)
			return
		}
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
	if !bindJSON(c, &in) {
		return
	}
	in.Pembeli = trim(in.Pembeli)
	if in.Tgl == "" || (in.Arah != "masuk" && in.Arah != "keluar") || in.Gram <= 0 || in.Harga < 0 {
		bad(c, "Tanggal, arah, jumlah, dan harga wajib diisi")
		return
	}
	if in.Arah == "keluar" && in.Pembeli == "" {
		bad(c, "Pembeli wajib diisi untuk keju keluar")
		return
	}
	if !validTgl(in.Tgl) || !textOK(in.Pembeli, 100) || !inRange(in.Gram, 0.01, maxGram) || !inRange(in.Harga, 0, maxUang) {
		bad(c, "Tanggal, jumlah, harga, atau nama pembeli tidak valid")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO melk_keju (tgl,arah,jumlah_gram,harga,pembeli) VALUES (?,?,?,?,NULLIF(?,''))",
			in.Tgl, in.Arah, in.Gram, in.Harga, in.Pembeli)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if in.Harga > 0 {
			arah, kat, ket := "keluar", "beli_keju", "Keju masuk dari kandang"
			if in.Arah == "keluar" {
				arah, kat, ket = "masuk", "jual_keju", "Jual keju ke "+in.Pembeli
			}
			if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
				VALUES (?,'melk',?,?,?,?,'melk_keju',?)`, in.Tgl, arah, kat, in.Harga, ket, id); err != nil {
				return err
			}
		}
		after, err := services.Snapshot(tx, "melk_keju", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "melk_keju", id, nil, after)
	})
	if done(c, "tambah keju", err) {
		ok(c)
	}
}

func DelKeju(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		row, err := services.Snapshot(tx, "melk_keju", id)
		if err != nil {
			return err
		}
		if row == nil {
			return errNotFound
		}
		kas, err := services.SnapshotWhere(tx, "kas", "ref_tabel='melk_keju' AND ref_id=?", id)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM kas WHERE ref_tabel='melk_keju' AND ref_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM melk_keju WHERE id=?", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "melk_keju", id, map[string]any{"melk_keju": row, "kas": kas}, nil)
	})
	if done(c, "hapus keju", err) {
		ok(c)
	}
}
