package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/models"
	"rejeki-baru-farm/internal/services"
)

var statusSapi = map[string]bool{"laktasi": true, "kering": true, "dara": true, "pedet": true}

func ListSapi(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,nomor,nama,tgl_lahir,jumlah_anak,status,
		TIMESTAMPDIFF(MONTH,tgl_lahir,CURDATE()) FROM sapi ORDER BY nomor`)
	if err != nil {
		fail(c, "list sapi", err)
		return
	}
	defer rows.Close()
	out := []models.Sapi{}
	for rows.Next() {
		var s models.Sapi
		if err := rows.Scan(&s.ID, &s.Nomor, &s.Nama, &s.TglLahir, &s.JumlahAnak, &s.Status, &s.UmurBulan); err != nil {
			fail(c, "list sapi", err)
			return
		}
		out = append(out, s)
	}
	c.JSON(200, out)
}

func SaveSapi(c *gin.Context) {
	// struct input khusus: hanya field ini yang bisa diisi dari client
	var in struct {
		Nomor      string `json:"nomor"`
		Nama       string `json:"nama"`
		TglLahir   string `json:"tgl_lahir"`
		JumlahAnak int    `json:"jumlah_anak"`
		Status     string `json:"status"`
	}
	if !bindJSON(c, &in) {
		return
	}
	in.Nomor, in.Nama = trim(in.Nomor), trim(in.Nama)
	if in.Nomor == "" || in.Nama == "" || in.TglLahir == "" || !statusSapi[in.Status] {
		bad(c, "Nomor, nama, tgl lahir, dan status wajib diisi")
		return
	}
	if !textOK(in.Nomor, 30) || !textOK(in.Nama, 100) {
		bad(c, "Nomor maksimal 30 karakter dan nama maksimal 100 karakter")
		return
	}
	if t, valid := parseTgl(in.TglLahir); !valid || t.Year() < 1990 || t.After(today().AddDate(0, 0, 1)) {
		bad(c, "Tanggal lahir tidak valid")
		return
	}
	if in.JumlahAnak < 0 || in.JumlahAnak > 30 {
		bad(c, "Jumlah anak harus antara 0 dan 30")
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
		if id > 0 {
			before, err := services.Snapshot(tx, "sapi", id)
			if err != nil {
				return err
			}
			if before == nil {
				return errNotFound
			}
			if _, err := tx.Exec("UPDATE sapi SET nomor=?,nama=?,tgl_lahir=?,jumlah_anak=?,status=? WHERE id=?",
				in.Nomor, in.Nama, in.TglLahir, in.JumlahAnak, in.Status, id); err != nil {
				return err
			}
			after, err := services.Snapshot(tx, "sapi", id)
			if err != nil {
				return err
			}
			return audit(tx, c, "ubah", "sapi", id, before, after)
		}
		res, err := tx.Exec("INSERT INTO sapi (nomor,nama,tgl_lahir,jumlah_anak,status) VALUES (?,?,?,?,?)",
			in.Nomor, in.Nama, in.TglLahir, in.JumlahAnak, in.Status)
		if err != nil {
			return err
		}
		nid, _ := res.LastInsertId()
		after, err := services.Snapshot(tx, "sapi", nid)
		if err != nil {
			return err
		}
		return audit(tx, c, "tambah", "sapi", nid, nil, after)
	})
	if isDuplicate(err) {
		bad(c, "Gagal simpan (nomor mungkin sudah ada)")
		return
	}
	if done(c, "simpan sapi", err) {
		ok(c)
	}
}

func DelSapi(c *gin.Context) {
	id, valid := paramID(c)
	if !valid {
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		before, err := services.Snapshot(tx, "sapi", id)
		if err != nil {
			return err
		}
		if before == nil {
			return errNotFound
		}
		if _, err := tx.Exec("DELETE FROM sapi WHERE id=?", id); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "sapi", id, before, nil)
	})
	if done(c, "hapus sapi", err) {
		ok(c)
	}
}
