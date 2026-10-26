package controllers

import (
	"log"

	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
)

// ---------- Rekap per 10 hari per pembeli ----------
func RekapMelk(c *gin.Context) {
	bulan, valid := bulanParam(c)
	if !valid {
		return
	}
	// nama tabel & kolom hanya dari konstanta di bawah (bukan dari input user)
	q := func(tabel, kolom string) []gin.H {
		rows, err := database.DB.Query(`SELECT LEAST(3,FLOOR((DAY(tgl)-1)/10)+1) p, IFNULL(pembeli,'-'), SUM(`+kolom+`), SUM(harga)
			FROM `+tabel+` WHERE arah='keluar' AND DATE_FORMAT(tgl,'%Y-%m')=? GROUP BY p, pembeli ORDER BY p, pembeli`, bulan)
		out := []gin.H{}
		if err != nil {
			log.Printf("rekap melk %s: %v", tabel, err)
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var p int
			var b string
			var j, t float64
			if err := rows.Scan(&p, &b, &j, &t); err != nil {
				log.Printf("rekap melk %s: %v", tabel, err)
				return out
			}
			out = append(out, gin.H{"periode": p, "pembeli": b, "jumlah": j, "total": t})
		}
		return out
	}
	c.JSON(200, gin.H{"susu": q("melk_susu", "liter"), "keju": q("melk_keju", "jumlah_gram")})
}
