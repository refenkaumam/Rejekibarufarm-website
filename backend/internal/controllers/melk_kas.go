package controllers

import (
	"github.com/gin-gonic/gin"
)

// ---------- Keuangan Melk ----------
func ListKasMelk(c *gin.Context) {
	bulan, valid := bulanParam(c)
	if !valid {
		return
	}
	listKasUnit(c, "melk", bulan)
}

func AddKasMelk(c *gin.Context) {
	addKas(c, "melk", map[string]map[string]bool{
		"masuk":  {"kas": true, "lainnya": true},
		"keluar": {"gaji": true, "operasional": true, "lainnya": true},
	})
}

func DelKasMelk(c *gin.Context) {
	if found, err := delKas(c, "melk"); err == nil && !found {
		bad(c, "Catatan otomatis dihapus dari menu asalnya (susu, keju, atau penjualan di kandang)")
	}
}
