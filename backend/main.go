package main

import (
	"log"

	"rejeki-baru-farm/internal/config"
	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/routes"
	"rejeki-baru-farm/internal/services"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "ganti-secret-ini" {
		log.Println("PERINGATAN: JWT_SECRET belum diatur, isi dengan nilai acak sebelum dipakai di server")
	}
	if err := database.Connect(cfg.DSN); err != nil {
		log.Fatal("DB tidak terhubung: ", err)
	}
	database.Migrate()
	services.InitAuth(cfg.JWTSecret)
	services.SeedAdmin()
	log.Fatal(routes.Setup().Run(":" + cfg.Port))
}
