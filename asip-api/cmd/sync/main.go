package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArminDashti/as-ip/server/internal/config"
	"github.com/ArminDashti/as-ip/server/internal/database"
	"github.com/ArminDashti/as-ip/server/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	service := sync.NewService(db, cfg)
	if err := service.Run(ctx); err != nil {
		log.Fatalf("sync failed: %v", err)
	}
}
