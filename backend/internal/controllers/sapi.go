package controllers

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/models"
)

func ListSapi(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id,nomor,nama,tgl_lahir,jumlah_anak,status,
		TIMESTAMPDIFF(MONTH,tgl_lahir,CURDATE()) FROM sapi ORDER BY nomor`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []models.Sapi{}
	for rows.Next() {
		var s models.Sapi
		rows.Scan(&s.ID, &s.Nomor, &s.Nama, &s.TglLahir, &s.JumlahAnak, &s.Status, &s.UmurBulan)
		out = append(out, s)
	}
	c.JSON(200, out)
}

func SaveSapi(c *gin.Context) {
	var s models.Sapi
	if c.BindJSON(&s) != nil {
		return
	}
	valid := map[string]bool{"laktasi": true, "kering": true, "dara": true, "pedet": true}
	if s.Nomor == "" || s.Nama == "" || s.TglLahir == "" || !valid[s.Status] {
		c.JSON(400, gin.H{"error": "Nomor, nama, tgl lahir, dan status wajib diisi"})
		return
	}
	var err error
	if id := c.Param("id"); id != "" {
		_, err = database.DB.Exec("UPDATE sapi SET nomor=?,nama=?,tgl_lahir=?,jumlah_anak=?,status=? WHERE id=?",
			s.Nomor, s.Nama, s.TglLahir, s.JumlahAnak, s.Status, id)
	} else {
		_, err = database.DB.Exec("INSERT INTO sapi (nomor,nama,tgl_lahir,jumlah_anak,status) VALUES (?,?,?,?,?)",
			s.Nomor, s.Nama, s.TglLahir, s.JumlahAnak, s.Status)
	}
	if err != nil {
		c.JSON(400, gin.H{"error": "Gagal simpan (nomor mungkin sudah ada)"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func DelSapi(c *gin.Context) {
	database.DB.Exec("DELETE FROM sapi WHERE id=?", c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}
