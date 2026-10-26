package controllers

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

// ---------- Kas manual (dipakai kandang, Melk, dan back office) ----------

type kasInput struct {
	Tgl        string  `json:"tgl"`
	Arah       string  `json:"arah"`
	Kategori   string  `json:"kategori"`
	Jumlah     float64 `json:"jumlah"`
	Keterangan string  `json:"keterangan"`
}

// addKas mencatat transaksi kas manual untuk satu unit. Unit dan kategori ditentukan server.
func addKas(c *gin.Context, unit string, kategori map[string]map[string]bool) {
	var in kasInput
	if !bindJSON(c, &in) {
		return
	}
	in.Keterangan = trim(in.Keterangan)
	if in.Tgl == "" || in.Jumlah <= 0 || !kategori[in.Arah][in.Kategori] {
		bad(c, "Tanggal, kategori, dan jumlah wajib diisi")
		return
	}
	if !validTgl(in.Tgl) || !inRange(in.Jumlah, 0.01, maxUang) || !textOK(in.Keterangan, 255) {
		bad(c, "Tanggal, jumlah, atau keterangan tidak valid")
		return
	}
	err := withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan) VALUES (?,?,?,?,?,?)`,
			in.Tgl, unit, in.Arah, in.Kategori, in.Jumlah, in.Keterangan)
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
	if done(c, "tambah kas "+unit, err) {
		ok(c)
	}
}

// delKas menghapus kas manual (bukan catatan otomatis) milik unit tersebut.
// Mengembalikan false kalau baris tidak ada / bukan milik unit / otomatis.
func delKas(c *gin.Context, unit string) (bool, error) {
	id, valid := paramID(c)
	if !valid {
		return true, nil
	}
	found := false
	err := withTx(func(tx *sql.Tx) error {
		rows, err := services.SnapshotWhere(tx, "kas", "id=? AND unit=? AND ref_tabel IS NULL", id, unit)
		if err != nil || len(rows) == 0 {
			return err
		}
		found = true
		if _, err := tx.Exec("DELETE FROM kas WHERE id=? AND unit=? AND ref_tabel IS NULL", id, unit); err != nil {
			return err
		}
		return audit(tx, c, "hapus", "kas", id, rows[0], nil)
	})
	if err != nil {
		fail(c, "hapus kas "+unit, err)
		return true, err
	}
	if found {
		ok(c)
	}
	return found, nil
}

// ---------- Keuangan kandang ----------
func ListKas(c *gin.Context) {
	bulan, valid := bulanParam(c)
	if !valid {
		return
	}
	listKasUnit(c, "kandang", bulan)
}

func listKasUnit(c *gin.Context, unit, bulan string) {
	rows, err := database.DB.Query(`SELECT id,tgl,arah,kategori,jumlah,IFNULL(keterangan,''),ref_tabel IS NOT NULL
		FROM kas WHERE unit=? AND DATE_FORMAT(tgl,'%Y-%m')=? ORDER BY tgl DESC,id DESC`, unit, bulan)
	if err != nil {
		fail(c, "list kas "+unit, err)
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, arah, kat, ket string
		var j float64
		var auto bool
		if err := rows.Scan(&id, &tgl, &arah, &kat, &j, &ket, &auto); err != nil {
			fail(c, "list kas "+unit, err)
			return
		}
		list = append(list, gin.H{"id": id, "tgl": tgl, "arah": arah, "kategori": kat, "jumlah": j, "keterangan": ket, "otomatis": auto})
	}
	var masuk, keluar, tm, tk float64
	if err := database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,0)),0),IFNULL(SUM(IF(arah='keluar',jumlah,0)),0)
		FROM kas WHERE unit=? AND DATE_FORMAT(tgl,'%Y-%m')=?`, unit, bulan).Scan(&masuk, &keluar); err != nil {
		fail(c, "list kas "+unit, err)
		return
	}
	if err := database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,0)),0),IFNULL(SUM(IF(arah='keluar',jumlah,0)),0)
		FROM kas WHERE unit=?`, unit).Scan(&tm, &tk); err != nil {
		fail(c, "list kas "+unit, err)
		return
	}
	c.JSON(200, gin.H{"bulan": bulan, "rows": list, "masuk": masuk, "keluar": keluar, "saldo": tm - tk})
}

func AddKas(c *gin.Context) {
	addKas(c, "kandang", map[string]map[string]bool{
		"masuk":  {"kas": true, "topup_bo": true, "lainnya": true},
		"keluar": {"gaji": true, "operasional": true, "lainnya": true},
	})
}

func DelKas(c *gin.Context) {
	if found, err := delKas(c, "kandang"); err == nil && !found {
		bad(c, "Catatan otomatis dihapus dari menu asalnya (pakan, penjualan, atau investor)")
	}
}
