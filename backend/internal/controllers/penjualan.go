package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// ---------- Penjualan susu ----------
func ListPenjualan(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,tgl,tujuan,liter,harga_per_liter,liter*harga_per_liter
		FROM penjualan_susu ORDER BY tgl DESC,id DESC LIMIT 100`)
	if err != nil {
		fail(c, "list penjualan", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, tujuan string
		var l, h, t float64
		if err := rows.Scan(&id, &tgl, &tujuan, &l, &h, &t); err != nil {
			fail(c, "list penjualan", err)
			return
		}
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
	if !bindJSON(c, &in) {
		return
	}
	if in.Tgl == "" || (in.Tujuan != "koperasi" && in.Tujuan != "melk_rbf") || in.Liter <= 0 || in.Harga < 0 {
		bad(c, "Tanggal, tujuan, liter, dan harga wajib diisi")
		return
	}
	if !validTgl(in.Tgl) {
		bad(c, "Tanggal tidak valid")
		return
	}
	total := in.Liter * in.Harga
	if !inRange(in.Liter, 0.01, maxLiter) || !inRange(in.Harga, 0, maxUang) || !inRange(total, 0, maxUang) {
		bad(c, "Liter atau harga di luar batas wajar")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO penjualan_susu (tgl,tujuan,liter,harga_per_liter) VALUES (?,?,?,?)", in.Tgl, in.Tujuan, in.Liter, in.Harga)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'kandang','masuk','pembelian_susu',?,?,'penjualan_susu',?)`, in.Tgl, total, "Jual susu ke "+in.Tujuan, id); err != nil {
			return err
		}
		if in.Tujuan == "melk_rbf" {
			if _, err := tx.Exec("INSERT INTO melk_susu (tgl,arah,liter,harga,ref_id) VALUES (?,'masuk',?,?,?)", in.Tgl, in.Liter, total, id); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
				VALUES (?,'melk','keluar','beli_susu',?,'Beli susu dari kandang','penjualan_susu',?)`, in.Tgl, total, id); err != nil {
				return err
			}
		}
		after, err := services.Snapshot(tx, "penjualan_susu", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "penjualan_susu", id, nil, after)
	})
	if done(c, "tambah penjualan", err) {
		ok(c)
	}
}

func DelPenjualan(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		row, err := services.Snapshot(tx, "penjualan_susu", id)
		if err != nil {
			return err
		}
		if row == nil {
			return errNotFound
		}
		kas, err := services.SnapshotWhere(tx, "kas", "ref_tabel='penjualan_susu' AND ref_id=?", id)
		if err != nil {
			return err
		}
		susu, err := services.SnapshotWhere(tx, "melk_susu", "ref_id=? AND arah='masuk'", id)
		if err != nil {
			return err
		}
		for _, q := range []string{
			"DELETE FROM kas WHERE ref_tabel='penjualan_susu' AND ref_id=?",
			"DELETE FROM melk_susu WHERE ref_id=? AND arah='masuk'",
			"DELETE FROM penjualan_susu WHERE id=?",
		} {
			if _, err := tx.Exec(q, id); err != nil {
				return err
			}
		}
		return audit(tx, c, "hapus", "penjualan_susu", id, map[string]any{"penjualan_susu": row, "kas": kas, "melk_susu": susu}, nil)
	})
	if done(c, "hapus penjualan", err) {
		ok(c)
	}
}
