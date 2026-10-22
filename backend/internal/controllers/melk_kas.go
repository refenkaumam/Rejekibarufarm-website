package controllers

import (
	"time"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Keuangan Melk ----------
func ListKasMelk(c *gin.Context) {
	bulan := c.Query("bulan")
	if bulan == "" {
		bulan = time.Now().Format("2006-01")
	}
	rows, err := database.DB.Query(`SELECT id,tgl,arah,kategori,jumlah,IFNULL(keterangan,''),ref_tabel IS NOT NULL
		FROM kas WHERE unit='melk' AND DATE_FORMAT(tgl,'%Y-%m')=? ORDER BY tgl DESC,id DESC`, bulan)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id int64
		var tgl, arah, kat, ket string
		var j float64
		var auto bool
		rows.Scan(&id, &tgl, &arah, &kat, &j, &ket, &auto)
		list = append(list, gin.H{"id": id, "tgl": tgl, "arah": arah, "kategori": kat, "jumlah": j, "keterangan": ket, "otomatis": auto})
	}
	var masuk, keluar, tm, tk float64
	database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,0)),0),IFNULL(SUM(IF(arah='keluar',jumlah,0)),0)
		FROM kas WHERE unit='melk' AND DATE_FORMAT(tgl,'%Y-%m')=?`, bulan).Scan(&masuk, &keluar)
	database.DB.QueryRow(`SELECT IFNULL(SUM(IF(arah='masuk',jumlah,0)),0),IFNULL(SUM(IF(arah='keluar',jumlah,0)),0)
		FROM kas WHERE unit='melk'`).Scan(&tm, &tk)
	c.JSON(200, gin.H{"bulan": bulan, "rows": list, "masuk": masuk, "keluar": keluar, "saldo": tm - tk})
}

func AddKasMelk(c *gin.Context) {
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
	database.DB.Exec(`INSERT INTO kas (tgl,unit,arah,kategori,jumlah,keterangan) VALUES (?,'melk',?,?,?,?)`,
		in.Tgl, in.Arah, in.Kategori, in.Jumlah, in.Keterangan)
	c.JSON(200, gin.H{"ok": true})
}

func DelKasMelk(c *gin.Context) {
	res, _ := database.DB.Exec("DELETE FROM kas WHERE id=? AND unit='melk' AND ref_tabel IS NULL", c.Param("id"))
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(400, gin.H{"error": "Catatan otomatis dihapus dari menu asalnya (susu, keju, atau penjualan di kandang)"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
