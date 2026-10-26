package controllers

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

func ListProduksi(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT tgl, SUM(IF(sesi='pagi',liter,0)), SUM(IF(sesi='sore',liter,0))
		FROM produksi_susu GROUP BY tgl ORDER BY tgl DESC LIMIT 31`)
	if err != nil {
		fail(c, "list produksi", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var t string
		var pg, sr float64
		if err := rows.Scan(&t, &pg, &sr); err != nil {
			fail(c, "list produksi", err)
			return
		}
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
	if !bindJSON(c, &in) {
		return
	}
	if in.Tgl == "" || (in.Sesi != "pagi" && in.Sesi != "sore") || in.Liter < 0 {
		bad(c, "Tanggal, sesi, dan liter wajib diisi")
		return
	}
	if !validTgl(in.Tgl) {
		bad(c, "Tanggal tidak valid")
		return
	}
	if !inRange(in.Liter, 0, maxLiter) {
		bad(c, "Jumlah liter di luar batas wajar")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		var id int64
		err := tx.QueryRow("SELECT id FROM produksi_susu WHERE tgl=? AND sesi=? FOR UPDATE", in.Tgl, in.Sesi).Scan(&id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var before map[string]any
		if id > 0 {
			if before, err = services.Snapshot(tx, "produksi_susu", id); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`INSERT INTO produksi_susu (tgl,sesi,liter) VALUES (?,?,?)
			ON DUPLICATE KEY UPDATE liter=VALUES(liter)`, in.Tgl, in.Sesi, in.Liter); err != nil {
			return err
		}
		if id == 0 {
			if err := tx.QueryRow("SELECT id FROM produksi_susu WHERE tgl=? AND sesi=?", in.Tgl, in.Sesi).Scan(&id); err != nil {
				return err
			}
		}
		after, err := services.Snapshot(tx, "produksi_susu", id)
		if err != nil {
			return err
		}
		aksi := "tambah"
		if before != nil {
			aksi = "ubah"
		}
		return audit(tx, c, aksi, "produksi_susu", id, before, after)
	})
	if done(c, "simpan produksi", err) {
		ok(c)
	}
}

func RekapProduksi(c *gin.Context) {
	f := "%Y-%m"
	if c.Query("by") == "tahun" {
		f = "%Y"
	}
	rows, err := database.DB.Query(`SELECT DATE_FORMAT(tgl,?) p, SUM(liter), COUNT(DISTINCT tgl)
		FROM produksi_susu GROUP BY p ORDER BY p DESC LIMIT 24`, f)
	if err != nil {
		fail(c, "rekap produksi", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var p string
		var t float64
		var h int
		if err := rows.Scan(&p, &t, &h); err != nil {
			fail(c, "rekap produksi", err)
			return
		}
		out = append(out, gin.H{"periode": p, "total": t, "hari": h})
	}
	c.JSON(200, out)
}
