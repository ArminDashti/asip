package repository

import (
	"context"
)

func (r *AsRepository) LogRequest(ctx context.Context, clientIP string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO api_request (client_ip) VALUES ($1)`, clientIP)
	return err
}
