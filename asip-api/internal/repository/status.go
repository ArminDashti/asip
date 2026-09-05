package repository

import (
	"context"
	"database/sql"
	"time"
)

type Status struct {
	LastSync          *time.Time
	RequestsToday     int64
	RequestsYesterday int64
}

func (r *AsRepository) GetStatus(ctx context.Context) (Status, error) {
	var status Status

	lastSync, err := r.GetLastSync(ctx)
	if err != nil {
		return status, err
	}
	status.LastSync = lastSync

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM api_request
		WHERE request_at::date = CURRENT_DATE
	`).Scan(&status.RequestsToday)
	if err != nil {
		return status, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM api_request
		WHERE request_at::date = CURRENT_DATE - INTERVAL '1 day'
	`).Scan(&status.RequestsYesterday)
	if err != nil {
		return status, err
	}

	return status, nil
}

func (r *AsRepository) GetLastSync(ctx context.Context) (*time.Time, error) {
	var lastSync sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT last_sync_at FROM sync_state WHERE id = 1`).Scan(&lastSync)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !lastSync.Valid {
		return nil, nil
	}
	syncedAt := lastSync.Time
	return &syncedAt, nil
}

func (r *AsRepository) SetLastSync(ctx context.Context, syncedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sync_state (id, last_sync_at) VALUES (1, $1)
		ON CONFLICT(id) DO UPDATE SET last_sync_at = excluded.last_sync_at
	`, syncedAt.UTC())
	return err
}
