package services

import (
	"time"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/models"
)

func MonthRange(bulan string) (string, string) {
	t, err := time.Parse("2006-01", bulan)
	if err != nil {
		t, _ = time.Parse("2006-01", time.Now().Format("2006-01"))
	}
	return t.Format("2006-01-02"), t.AddDate(0, 1, -1).Format("2006-01-02")
}

// Laba rugi gabungan 3 unit. Penjualan susu kandang ke Melk dieliminasi
// (pendapatan kandang = biaya beli susu Melk), setoran kas, top up, dan modal investor bukan laba.
func HitungLR(dari, sampai string) models.LR {
	m := map[string]float64{}
	rows, err := database.DB.Query(`SELECT unit,arah,kategori,SUM(jumlah) FROM kas WHERE tgl BETWEEN ? AND ? GROUP BY unit,arah,kategori`, dari, sampai)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u, a, k string
			var s float64
			rows.Scan(&u, &a, &k, &s)
			m[u+"/"+a+"/"+k] = s
		}
	}
	all := func(a, k string) float64 {
		return m["kandang/"+a+"/"+k] + m["melk/"+a+"/"+k] + m["back_office/"+a+"/"+k]
	}
	elim := m["melk/keluar/beli_susu"]
	p := []models.Baris{
		{Label: "Penjualan susu ke koperasi", Jumlah: m["kandang/masuk/pembelian_susu"] - elim},
		{Label: "Penjualan susu oleh Melk by RBF", Jumlah: m["melk/masuk/jual_susu"]},
		{Label: "Penjualan keju", Jumlah: m["melk/masuk/jual_keju"]},
		{Label: "Pendapatan lain", Jumlah: all("masuk", "lainnya")},
	}
	b := []models.Baris{
		{Label: "Pembelian pakan", Jumlah: m["kandang/keluar/beli_pakan"]},
		{Label: "Bulanan investor", Jumlah: m["kandang/keluar/bulanan_investor"]},
		{Label: "Pembelian keju", Jumlah: m["melk/keluar/beli_keju"]},
		{Label: "Gaji pegawai", Jumlah: all("keluar", "gaji")},
		{Label: "Operasional", Jumlah: all("keluar", "operasional")},
		{Label: "Biaya lain", Jumlah: all("keluar", "lainnya")},
	}
	r := models.LR{Pendapatan: p, Biaya: b}
	for _, x := range p {
		r.TotalPendapatan += x.Jumlah
	}
	for _, x := range b {
		r.TotalBiaya += x.Jumlah
	}
	r.Laba = r.TotalPendapatan - r.TotalBiaya
	return r
}

func KasAwal() (string, float64) {
	tgl, j := "1970-01-01", 0.0
	database.DB.QueryRow("SELECT tgl,jumlah FROM kas_awal_bo ORDER BY tgl DESC,id DESC LIMIT 1").Scan(&tgl, &j)
	return tgl, j
}

// Saldo BO = kas awal + kas BO masuk - kas BO keluar - top up ke kandang (sejak tanggal kas awal)
func SaldoBO() float64 {
	tgl, j := KasAwal()
	var s, topup float64
	database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,-jumlah)),0) FROM kas WHERE unit='back_office' AND tgl>=?`, tgl).Scan(&s)
	database.DB.QueryRow(`SELECT IFNULL(SUM(jumlah),0) FROM kas WHERE unit='kandang' AND kategori='topup_bo' AND tgl>=?`, tgl).Scan(&topup)
	return j + s - topup
}
