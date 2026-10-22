package controllers

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/services"
)

func rangeQuery(c *gin.Context) (string, string) {
	d, s := services.MonthRange(time.Now().Format("2006-01"))
	if v := c.Query("dari"); v != "" {
		d = v
	}
	if v := c.Query("sampai"); v != "" {
		s = v
	}
	return d, s
}

func BoRingkasan(c *gin.Context) {
	saldo := func(unit string) float64 {
		var s float64
		database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,-jumlah)),0) FROM kas WHERE unit=?`, unit).Scan(&s)
		return s
	}
	bo, kd, ml := services.SaldoBO(), saldo("kandang"), saldo("melk")
	d, s := services.MonthRange(time.Now().Format("2006-01"))
	lr := services.HitungLR(d, s)
	c.JSON(200, gin.H{"saldo_bo": bo, "saldo_kandang": kd, "saldo_melk": ml, "saldo_total": bo + kd + ml,
		"laba_bulan_ini": lr.Laba})
}

func BoLabaRugi(c *gin.Context) {
	d, s := rangeQuery(c)
	c.JSON(200, gin.H{"dari": d, "sampai": s, "lr": services.HitungLR(d, s)})
}

func BoKasAwal(c *gin.Context) {
	t, j := services.KasAwal()
	c.JSON(200, gin.H{"tgl": t, "jumlah": j})
}

func BoSetKasAwal(c *gin.Context) {
	var in struct {
		Tgl    string  `json:"tgl"`
		Jumlah float64 `json:"jumlah"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	if in.Tgl == "" || in.Jumlah < 0 {
		c.JSON(400, gin.H{"error": "Tanggal dan jumlah kas awal wajib diisi"})
		return
	}
	database.DB.Exec("INSERT INTO kas_awal_bo (tgl,jumlah) VALUES (?,?)", in.Tgl, in.Jumlah)
	c.JSON(200, gin.H{"ok": true})
}

func BoListKas(c *gin.Context) {
	bulan := c.Query("bulan")
	if bulan == "" {
		bulan = time.Now().Format("2006-01")
	}
	rows, err := database.DB.Query(`
		SELECT id,tgl,arah,kategori,jumlah,IFNULL(keterangan,''),0 FROM kas WHERE unit='back_office' AND DATE_FORMAT(tgl,'%Y-%m')=?
		UNION ALL
		SELECT id,tgl,'keluar','topup_kandang',jumlah,'Top up ke kandang',1 FROM kas WHERE unit='kandang' AND kategori='topup_bo' AND DATE_FORMAT(tgl,'%Y-%m')=?
		ORDER BY tgl DESC,id DESC`, bulan, bulan)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := []gin.H{}
	var masuk, keluar float64
	for rows.Next() {
		var id int64
		var tgl, arah, kat, ket string
		var j float64
		var auto bool
		rows.Scan(&id, &tgl, &arah, &kat, &j, &ket, &auto)
		if arah == "masuk" {
			masuk += j
		} else {
			keluar += j
		}
		list = append(list, gin.H{"id": id, "tgl": tgl, "arah": arah, "kategori": kat, "jumlah": j, "keterangan": ket, "otomatis": auto})
	}
	c.JSON(200, gin.H{"rows": list, "masuk": masuk, "keluar": keluar, "saldo": services.SaldoBO()})
}

func BoAddKas(c *gin.Context) {
	var in struct {
		Tgl        string  `json:"tgl"`
		Arah       string  `json:"arah"`
		Kategori   string  `json:"kategori"`
		Jumlah     float64 `json:"jumlah"`
		Keterangan string  `json:"keterangan"`
	}
	if c.BindJSON(&in) != nil {
		return
	}
	ok := map[string]map[string]bool{
		"masuk":  {"kas": true, "lainnya": true},
		"keluar": {"gaji": true, "operasional": true, "lainnya": true},
	}
	if in.Tgl == "" || in.Jumlah <= 0 || !ok[in.Arah][in.Kategori] {
		c.JSON(400, gin.H{"error": "Tanggal, kategori, dan jumlah wajib diisi"})
		return
	}
	database.DB.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan) VALUES (?,'back_office',?,?,?,?)`,
		in.Tgl, in.Arah, in.Kategori, in.Jumlah, in.Keterangan)
	c.JSON(200, gin.H{"ok": true})
}

func BoDelKas(c *gin.Context) {
	database.DB.Exec("DELETE FROM kas WHERE id=? AND unit='back_office'", c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}

// Export CSV (buka di Excel): laba rugi + semua transaksi kas pada rentang tanggal
func BoExport(c *gin.Context) {
	d, s := rangeQuery(c)
	lr := services.HitungLR(d, s)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=laporan_"+d+"_"+s+".csv")
	c.Writer.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(c.Writer)
	w.Comma = ';'
	f := func(v float64) string { return strings.TrimRight(strings.TrimRight(strconvF(v), "0"), ".") }
	w.Write([]string{"Rejeki Baru Farm - Laporan Laba Rugi"})
	w.Write([]string{"Periode", d, "sampai", s})
	w.Write([]string{})
	w.Write([]string{"PENDAPATAN", ""})
	for _, x := range lr.Pendapatan {
		w.Write([]string{x.Label, f(x.Jumlah)})
	}
	w.Write([]string{"Total pendapatan", f(lr.TotalPendapatan)})
	w.Write([]string{"BIAYA", ""})
	for _, x := range lr.Biaya {
		w.Write([]string{x.Label, f(x.Jumlah)})
	}
	w.Write([]string{"Total biaya", f(lr.TotalBiaya)})
	w.Write([]string{"LABA / (RUGI)", f(lr.Laba)})
	w.Write([]string{})
	w.Write([]string{"TRANSAKSI KAS"})
	w.Write([]string{"Tanggal", "Unit", "Arah", "Kategori", "Jumlah", "Keterangan"})
	rows, err := database.DB.Query(`SELECT tgl,unit,arah,kategori,jumlah,IFNULL(keterangan,'') FROM kas WHERE tgl BETWEEN ? AND ? ORDER BY tgl,id`, d, s)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tgl, u, a, k, ket string
			var j float64
			rows.Scan(&tgl, &u, &a, &k, &j, &ket)
			w.Write([]string{tgl, u, a, k, f(j), ket})
		}
	}
	w.Flush()
}

func strconvF(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
