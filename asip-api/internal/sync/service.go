package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

type Service struct {
	db  *sql.DB
	cfg *config.Config
}

func NewService(db *sql.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) Run(ctx context.Context) error {
	startedAt := time.Now()
	log.Println("sync: starting daily data sync")

	repos := []struct {
		url  string
		path string
		name string
	}{
		{s.cfg.URLAsMetadata, s.cfg.RepoAsMetadata, "as-metadata"},
		{s.cfg.URLAsIPBlocks, s.cfg.RepoAsIPBlocks, "as-ip-blocks"},
		{s.cfg.URLGeoIPBlocks, s.cfg.RepoGeoIPBlocks, "geo-ip-blocks"},
	}

	for _, repo := range repos {
		if err := cloneOrPull(ctx, repo.url, repo.path); err != nil {
			return fmt.Errorf("update repo %s: %w", repo.name, err)
		}
	}

	log.Println("sync: purging existing dataset")
	if err := purgeAllData(ctx, s.db); err != nil {
		return err
	}

	log.Println("sync: importing as-metadata")
	if err := importMetadata(ctx, s.db, s.cfg.RepoAsMetadata); err != nil {
		return fmt.Errorf("import metadata: %w", err)
	}

	log.Println("sync: ensuring geo countries exist")
	if err := ensureGeoCountries(ctx, s.db, s.cfg.RepoGeoIPBlocks); err != nil {
		return fmt.Errorf("ensure geo countries: %w", err)
	}

	log.Println("sync: importing as-ip-blocks prefixes")
	if err := importAsPrefixes(ctx, s.db, s.cfg.RepoAsIPBlocks); err != nil {
		return fmt.Errorf("import as prefixes: %w", err)
	}

	log.Println("sync: importing geo-ip-blocks prefixes")
	if err := importGeoPrefixes(ctx, s.db, s.cfg.RepoGeoIPBlocks); err != nil {
		return fmt.Errorf("import geo prefixes: %w", err)
	}

	completedAt := time.Now()
	if err := repository.NewAsRepository(s.db).SetLastSync(ctx, completedAt); err != nil {
		return fmt.Errorf("record last sync: %w", err)
	}

	log.Printf("sync: completed in %s", time.Since(startedAt).Round(time.Second))
	return nil
}
