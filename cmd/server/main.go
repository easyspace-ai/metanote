package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"youmind-backend-v2/internal/config"
	"youmind-backend-v2/internal/infrastructure/persistence"
	"youmind-backend-v2/internal/interfaces/http"
)

func main() {
	cfg := config.Load()
	db, err := persistence.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	router := http.Wire(cfg, db)
	srvAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("🚀 YouMind Backend v2 (DDD + GORM) on %s (env=%s)", srvAddr, cfg.AppEnv)

	go func() {
		if err := router.Run(srvAddr); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("👋 Shutting down...")
}
