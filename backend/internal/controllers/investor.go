package controllers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Investor ----------
func ListInvestor(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT i.id,i.nama,IFNULL(i.kontak,''),i.jumlah_invest,i.mulai,i.selesai,
		IFNULL(GROUP_CONCAT(s.nomor ORDER BY s.nomor SEPARATOR ', '),''),
		IFNULL((SELECT SUM(jumlah) FROM kas WHERE ref_tabel='investor' AND ref_id=i.id),0)
		FROM investor i LEFT JOIN sapi s ON s.investor_id=i.id GROUP BY i.id ORDER BY i.nama`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var nama, kontak, mulai, selesai, sapi string
		var jml, dibayar float64
		rows.Scan(&id, &nama, &kontak, &jml, &mulai, &selesai, &sapi, &dibayar)
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
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Nama == "" || in.Jumlah <= 0 || in.Mulai == "" || in.Selesai == "" {
		c.JSON(400, gin.H{"error": "Nama, jumlah invest, dan periode wajib diisi"})
		return
	}
	tx, _ := database.DB.Begin()
	id := c.Param("id")
	if id == "" {
		res, err := tx.Exec("INSERT INTO investor (nama,kontak,jumlah_invest,mulai,selesai) VALUES (?,?,?,?,?)",
			in.Nama, in.Kontak, in.Jumlah, in.Mulai, in.Selesai)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Gagal simpan"})
			return
		}
		n, _ := res.LastInsertId()
		id = strconv.FormatInt(n, 10)
		tx.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
			VALUES (?,'kandang','masuk','investor',?,?,'investor_modal',?)`, in.Mulai, in.Jumlah, "Modal investor "+in.Nama, id)
	} else {
		tx.Exec("UPDATE investor SET nama=?,kontak=?,jumlah_invest=?,mulai=?,selesai=? WHERE id=?",
			in.Nama, in.Kontak, in.Jumlah, in.Mulai, in.Selesai, id)
		tx.Exec("UPDATE kas SET tgl=?,jumlah=?,keterangan=? WHERE ref_tabel='investor_modal' AND ref_id=?",
			in.Mulai, in.Jumlah, "Modal investor "+in.Nama, id)
	}
	tx.Exec("UPDATE sapi SET investor_id=NULL WHERE investor_id=?", id)
	for _, no := range strings.Split(in.NomorSapi, ",") {
		if no = strings.TrimSpace(no); no != "" {
			tx.Exec("UPDATE sapi SET investor_id=? WHERE nomor=?", id, no)
		}
	}
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func DelInvestor(c *gin.Context) {
	tx, _ := database.DB.Begin()
	tx.Exec("UPDATE sapi SET investor_id=NULL WHERE investor_id=?", c.Param("id"))
	tx.Exec("DELETE FROM kas WHERE ref_tabel IN ('investor','investor_modal') AND ref_id=?", c.Param("id"))
	tx.Exec("DELETE FROM investor WHERE id=?", c.Param("id"))
	tx.Commit()
	c.JSON(200, gin.H{"ok": true})
}

func ListBayar(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT k.id,k.tgl,i.nama,k.jumlah,IFNULL(k.keterangan,'')
		FROM kas k JOIN investor i ON i.id=k.ref_id WHERE k.ref_tabel='investor'
		ORDER BY k.tgl DESC,k.id DESC LIMIT 100`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, nama, ket string
		var j float64
		rows.Scan(&id, &tgl, &nama, &j, &ket)
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
	if c.BindJSON(&in) != nil {
		return
	}
	if in.InvestorID == 0 || in.Tgl == "" || in.Jumlah <= 0 {
		c.JSON(400, gin.H{"error": "Investor, tanggal, dan jumlah wajib diisi"})
		return
	}
	database.DB.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id)
		VALUES (?,'kandang','keluar','bulanan_investor',?,?,'investor',?)`, in.Tgl, in.Jumlah, in.Keterangan, in.InvestorID)
	c.JSON(200, gin.H{"ok": true})
}

func DelBayar(c *gin.Context) {
	database.DB.Exec("DELETE FROM kas WHERE id=? AND ref_tabel='investor'", c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}
