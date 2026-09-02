package sync

import (
	"context"
	"database/sql"
	"fmt"
)

func purgeAllData(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
DELETE FROM connectivity;
DELETE FROM asn_ipv6;
DELETE FROM asn_ipv4;
DELETE FROM country_ipv6;
DELETE FROM country_ipv4;
DELETE FROM ipv6_stat;
DELETE FROM ipv4_stat;
DELETE FROM asn;
DELETE FROM role;
DELETE FROM category;
DELETE FROM origin;
DELETE FROM country;
`)
	if err != nil {
		return fmt.Errorf("purge data: %w", err)
	}
	return nil
}
