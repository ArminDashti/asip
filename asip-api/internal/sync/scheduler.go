package sync

import (
	"context"
	"log"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
)

type Scheduler struct {
	service *Service
	cfg     *config.Config
}

func NewScheduler(service *Service, cfg *config.Config) *Scheduler {
	return &Scheduler{service: service, cfg: cfg}
}

func (s *Scheduler) Start(ctx context.Context) {
	if !s.cfg.SyncEnabled {
		log.Println("sync: scheduler disabled (SYNC_ENABLED=false)")
		return
	}

	go s.run(ctx)
}

func (s *Scheduler) run(ctx context.Context) {
	if s.cfg.SyncOnStartup {
		s.execute(ctx)
	}

	for {
		wait := durationUntilHourUTC(s.cfg.SyncHourUTC)
		log.Printf("sync: next run in %s (at %02d:00 UTC)", wait.Round(time.Minute), s.cfg.SyncHourUTC)

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			s.execute(ctx)
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}
}

func (s *Scheduler) execute(ctx context.Context) {
	runCtx, cancel := context.WithTimeout(ctx, 6*time.Hour)
	defer cancel()

	if err := s.service.Run(runCtx); err != nil {
		log.Printf("sync: failed: %v", err)
	}
}

func durationUntilHourUTC(hour int) time.Duration {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}
