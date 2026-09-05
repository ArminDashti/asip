package sync

import (
	"context"
	"log"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
	"github.com/ArminDashti/as-ip/server/internal/repository"
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
		wait := s.durationUntilNextSync(ctx)
		log.Printf(
			"sync: next run in %s (every %d day(s) at %02d:00 UTC)",
			wait.Round(time.Minute),
			s.cfg.SyncIntervalDays,
			s.cfg.SyncHourUTC,
		)

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

func (s *Scheduler) durationUntilNextSync(ctx context.Context) time.Duration {
	lastSync, err := repository.NewAsRepository(s.service.db).GetLastSync(ctx)
	if err != nil {
		log.Printf("sync: read last sync failed, scheduling for next sync hour: %v", err)
		lastSync = nil
	}
	return durationUntilNextSync(lastSync, s.cfg.SyncHourUTC, s.cfg.SyncIntervalDays, time.Now().UTC())
}

// durationUntilNextSync returns how long to wait before the next purge+import.
// The dataset is fully replaced on each run so only the latest IP version remains.
func durationUntilNextSync(lastSync *time.Time, hourUTC, intervalDays int, now time.Time) time.Duration {
	now = now.UTC()
	interval := time.Duration(intervalDays) * 24 * time.Hour

	var due time.Time
	if lastSync == nil {
		due = nextOccurrenceOfHourUTC(now, hourUTC)
	} else {
		earliest := lastSync.UTC().Add(interval)
		due = time.Date(earliest.Year(), earliest.Month(), earliest.Day(), hourUTC, 0, 0, 0, time.UTC)
		if due.Before(earliest) {
			due = due.Add(24 * time.Hour)
		}
		if !due.After(now) {
			due = nextOccurrenceOfHourUTC(now, hourUTC)
		}
	}

	wait := due.Sub(now)
	if wait < time.Minute {
		return time.Minute
	}
	return wait
}

func nextOccurrenceOfHourUTC(now time.Time, hourUTC int) time.Time {
	now = now.UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hourUTC, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
