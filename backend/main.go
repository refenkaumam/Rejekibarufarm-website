package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rejeki-baru-farm/internal/config"
	"rejeki-baru-farm/internal/database"
	"rejeki-baru-farm/internal/routes"
	"rejeki-baru-farm/internal/services"
)

func main() {
	cfg := config.Load()
	if err := database.Connect(cfg.DSN); err != nil {
		log.Fatal("DB tidak terhubung: ", err)
	}
	if err := database.Migrate(); err != nil {
		log.Fatal("Migrasi database gagal: ", err)
	}
	services.InitAuth(services.AuthOptions{
		Secret: cfg.JWTSecret, TTL: cfg.SessionTTL,
		CookieSecure: cfg.CookieSecure, CookieSameSite: cfg.CookieSameSite,
	})
	if err := services.SeedAdmin(cfg.AdminPassword); err != nil {
		log.Fatal("Gagal membuat akun awal: ", err)
	}

	srv := &http.Server{
		Addr:              cfg.Host + ":" + cfg.Port,
		Handler:           routes.Setup(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	log.Printf("API berjalan di %s:%s (mode %s)", cfg.Host, cfg.Port, cfg.Env)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
