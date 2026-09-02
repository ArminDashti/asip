package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
	"github.com/ArminDashti/as-ip/server/internal/database"
	"github.com/ArminDashti/as-ip/server/internal/handler"
	"github.com/ArminDashti/as-ip/server/internal/metrics"
	"github.com/ArminDashti/as-ip/server/internal/repository"
	"github.com/ArminDashti/as-ip/server/internal/router"
	"github.com/ArminDashti/as-ip/server/internal/service"
	"github.com/ArminDashti/as-ip/server/internal/sync"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(shutdownCtx, cfg.DB)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	asRepository := repository.NewAsRepository(pool)
	lookupService := service.NewLookupService(asRepository)

	syncService := sync.NewService(pool, cfg)
	sync.NewScheduler(syncService, cfg).Start(shutdownCtx)

	healthHandler := handler.NewHealthHandler()
	docsHandler := handler.NewDocsHandler()
	lookupHandler := handler.NewLookupHandler(lookupService)
	headersHandler := handler.NewHeadersHandler()

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	metrics.Register(engine)
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	// Trust private/Docker networks so ClientIP() uses X-Forwarded-For from HAProxy.
	if err := engine.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.1/32",
		"::1/128",
	}); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}
	router.Register(engine, healthHandler, docsHandler, lookupHandler, headersHandler, asRepository)

	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("as-ip server listening on :%d", cfg.Port)
		if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			serverErr <- serveErr
		}
	}()

	select {
	case err := <-serverErr:
		log.Fatalf("listen: %v", err)
	case <-shutdownCtx.Done():
	}
	log.Println("shutting down server")

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(gracefulCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
