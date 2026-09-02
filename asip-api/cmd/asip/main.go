package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/ArminDashti/as-ip/server/internal/config"
	"github.com/ArminDashti/as-ip/server/internal/database"
	"github.com/ArminDashti/as-ip/server/internal/repository"
	"github.com/ArminDashti/as-ip/server/internal/sync"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "status":
		if err := runStatus(); err != nil {
			log.Fatalf("status: %v", err)
		}
	case "sync":
		if err := runSync(); err != nil {
			log.Fatalf("sync: %v", err)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: asip <command>

Commands:
  status    Show last sync time and request counts
  sync      Run a one-off data sync`)
}

func runStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()

	status, err := repository.NewAsRepository(db).GetStatus(ctx)
	if err != nil {
		return err
	}

	lastSync := "never"
	if status.LastSync != nil {
		lastSync = status.LastSync.Local().Format("2006-01-02 15:04:05")
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "LAST-SYNC\tREQUESTS TODAY\tREQUESTS YESTERDAY")
	fmt.Fprintf(writer, "%s\t%d\t%d\n", lastSync, status.RequestsToday, status.RequestsYesterday)
	return writer.Flush()
}

func runSync() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()

	return sync.NewService(db, cfg).Run(ctx)
}
