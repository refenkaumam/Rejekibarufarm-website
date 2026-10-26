package controllers

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// ---------- Investor ----------
func ListInvestor(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT i.id,i.nama,IFNULL(i.kontak,''),i.jumlah_invest,i.mulai,i.selesai,
		IFNULL(GROUP_CONCAT(s.nomor ORDER BY s.nomor SEPARATOR ', '),''),
		IFNULL((SELECT SUM(jumlah) FROM kas WHERE ref_tabel='investor' AND ref_id=i.id),0)
		FROM investor i LEFT JOIN sapi s ON s.investor_id=i.id GROUP BY i.id ORDER BY i.nama`)
	if err != nil {
		fail(c, "list investor", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var nama, kontak, mulai, selesai, sapi string
		var jml, dibayar float64
		if err := rows.Scan(&id, &nama, &kontak, &jml, &mulai, &selesai, &sapi, &dibayar); err != nil {
			fail(c, "list investor", err)
			return
		}
		out = append(out, gin.H{"id": id, "nama": nama, "kontak": kontak, "jumlah_invest": jml,
			"mulai": mulai, "selesai": selesai, "nomor_sapi": sapi, "total_dibayar": dibayar})
	}
	c.JSON(200, out)
}

// Investor baru -> kas masuk kandang (kategori investor). Nomor sapi dipisah koma.
func SaveInvestor(c *gin.Context) {
	var in struct {
		Nama      string  `json:"nama"`
		Kontak    string  `json:"kontak"`
		Jumlah    float64 `json:"jumlah_invest"`
		Mulai     string  `json:"mulai"`
		Selesai   string  `json:"selesai"`
		NomorSapi string  `json:"nomor_sapi"`
	}
	if !bindJSON(c, &in) {
		return
	}
	in.Nama, in.Kontak = trim(in.Nama), trim(in.Kontak)
	if in.Nama == "" || in.Jumlah <= 0 || in.Mulai == "" || in.Selesai == "" {
		bad(c, "Nama, jumlah invest, dan periode wajib diisi")
		return
	}
	mulai, okM := parseTgl(in.Mulai)
	selesai, okS := parseTgl(in.Selesai)
	switch {
	case !textOK(in.Nama, 100) || !textOK(in.Kontak, 100):
		bad(c, "Nama dan kontak maksimal 100 karakter")
		return
	case !inRange(in.Jumlah, 1, maxUang):
		bad(c, "Jumlah invest di luar batas wajar")
		return
	case !okM || !okS || !validTgl(in.Mulai) || selesai.Before(mulai) || selesai.After(mulai.AddDate(50, 0, 0)):
		bad(c, "Periode kerja sama tidak valid")
		return
	case !textOK(in.NomorSapi, 2000):
		bad(c, "Daftar nomor sapi terlalu panjang")
		return
	}
	var id int64
	if c.Param("id") != "" {
		var valid bool
		if id, valid = paramID(c); !valid {
			return
		}
	}
	err := withTx(func(tx *sql.Tx) error {
		var before map[string]any
		if id == 0 {
			res, err := tx.Exec("INSERT INTO investor (nama,kontak,jumlah_invest,mulai,selesai) VALUES (?,?,?,?,?)",
				in.Nama, in.Kontak, in.Jumlah, in.Mulai, in.Selesai)
			if err != nil {
				return err
			}
			id, _ = res.LastInsertId()
			if _, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
				VALUES (?,'kandang','masuk','investor',?,?,'investor_modal',?)`, in.Mulai, in.Jumlah, "Modal investor "+in.Nama, id); err != nil {
				return err
			}
		} else {
			var err error
			if before, err = services.Snapshot(tx, "investor", id); err != nil {
				return err
			}
			if before == nil {
				return errNotFound
			}
			if _, err := tx.Exec("UPDATE investor SET nama=?,kontak=?,jumlah_invest=?,mulai=?,selesai=? WHERE id=?",
				in.Nama, in.Kontak, in.Jumlah, in.Mulai, in.Selesai, id); err != nil {
				return err
			}
			if _, err := tx.Exec("UPDATE kas SET tgl=?,jumlah=?,keterangan=? WHERE ref_tabel='investor_modal' AND ref_id=?",
				in.Mulai, in.Jumlah, "Modal investor "+in.Nama, id); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("UPDATE sapi SET investor_id=NULL WHERE investor_id=?", id); err != nil {
			return err
		}
		for _, no := range strings.Split(in.NomorSapi, ",") {
			if no = strings.TrimSpace(no); no != "" {
				if _, err := tx.Exec("UPDATE sapi SET investor_id=? WHERE nomor=?", id, no); err != nil {
					return err
				}
			}
		}
		after, err := services.Snapshot(tx, "investor", id)
		if err != nil {
			return err
		}
		after["nomor_sapi"] = in.NomorSapi
		aksi := "tambah"
		if before != nil {
			aksi = "ubah"
		}
		return audit(tx, c, aksi, "investor", id, before, after)
	})
	if done(c, "simpan investor", err) {
		ok(c)
	}
}

func DelInvestor(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		row, err := services.Snapshot(tx, "investor", id)
		if err != nil {
			return err
		}
		if row == nil {
			return errNotFound
		}
		kas, err := services.SnapshotWhere(tx, "kas", "ref_tabel IN ('investor','investor_modal') AND ref_id=?", id)
		if err != nil {
			return err
		}
		for _, q := range []string{
			"UPDATE sapi SET investor_id=NULL WHERE investor_id=?",
			"DELETE FROM kas WHERE ref_tabel IN ('investor','investor_modal') AND ref_id=?",
			"DELETE FROM investor WHERE id=?",
		} {
			if _, err := tx.Exec(q, id); err != nil {
				return err
			}
		}
		return audit(tx, c, "hapus", "investor", id, map[string]any{"investor": row, "kas": kas}, nil)
	})
	if done(c, "hapus investor", err) {
		ok(c)
	}
}

func ListBayar(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT k.id,k.tgl,i.nama,k.jumlah,IFNULL(k.keterangan,'')
		FROM kas k JOIN investor i ON i.id=k.ref_id WHERE k.ref_tabel='investor'
		ORDER BY k.tgl DESC,k.id DESC LIMIT 100`)
	if err != nil {
		fail(c, "list bayar investor", err)
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, nama, ket string
		var j float64
		if err := rows.Scan(&id, &tgl, &nama, &j, &ket); err != nil {
			fail(c, "list bayar investor", err)
			return
		}
		out = append(out, gin.H{"id": id, "tgl": tgl, "nama": nama, "jumlah": j, "keterangan": ket})
	}
	c.JSON(200, out)
}

// Bayar bulanan investor -> kas keluar kandang
func AddBayar(c *gin.Context) {
	var in struct {
		InvestorID int64   `json:"investor_id"`
		Tgl        string  `json:"tgl"`
		Jumlah     float64 `json:"jumlah"`
		Keterangan string  `json:"keterangan"`
	}
	if !bindJSON(c, &in) {
		return
	}
	in.Keterangan = trim(in.Keterangan)
	if in.InvestorID <= 0 || in.Tgl == "" || in.Jumlah <= 0 {
		bad(c, "Investor, tanggal, dan jumlah wajib diisi")
		return
	}
	if !validTgl(in.Tgl) || !inRange(in.Jumlah, 0.01, maxUang) || !textOK(in.Keterangan, 255) {
		bad(c, "Tanggal, jumlah, atau keterangan tidak valid")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		inv, err := services.Snapshot(tx, "investor", in.InvestorID)
		if err != nil {
			return err
		}
		if inv == nil {
			return errBad("Investor tidak ditemukan")
		}
		res, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'kandang','keluar','bulanan_investor',?,?,'investor',?)`, in.Tgl, in.Jumlah, in.Keterangan, in.InvestorID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		after, err := services.Snapshot(tx, "kas", id)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "kas", id, nil, after)
	})
	if done(c, "bayar investor", err) {
		ok(c)
	}
}

func DelBayar(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		rows, err := services.SnapshotWhere(tx, "kas", "id=? AND ref_tabel='investor'", id)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return errNotFound
		}
		if _, err := tx.Exec("DELETE FROM kas WHERE id=? AND ref_tabel='investor'", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "kas", id, rows[0], nil)
	})
	if done(c, "hapus bayar investor", err) {
		ok(c)
	}
}
