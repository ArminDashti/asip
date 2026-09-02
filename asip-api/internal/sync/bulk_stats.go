package sync

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

func insertIpv4Stats(ctx context.Context, db *sql.DB, records []asnRow, asnIDByNumber map[int]int) error {
	rows := make([][]any, 0, len(records))
	for _, record := range records {
		asnID, ok := asnIDByNumber[record.Number]
		if !ok {
			continue
		}
		rows = append(rows, []any{
			asnID,
			record.Ipv4Prefixes,
			record.Ipv4PrefixesAgg,
			record.Ipv4LargestPrefix,
			record.Ipv4TotalAddresses,
		})
	}

	for start := 0; start < len(rows); start += metadataBatchSize {
		end := start + metadataBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[start:end]
		if len(batch) == 0 {
			continue
		}
		if err := database.BulkInsert(ctx, db, "ipv4_stat",
			[]string{"asn_id", "prefixes", "prefixes_aggregated", "largest_prefix", "total_addresses"},
			batch,
		); err != nil {
			return fmt.Errorf("insert ipv4_stat batch %d: %w", start, err)
		}
	}

	return nil
}

func insertIpv6Stats(ctx context.Context, db *sql.DB, records []asnRow, asnIDByNumber map[int]int) error {
	rows := make([][]any, 0, len(records))
	for _, record := range records {
		asnID, ok := asnIDByNumber[record.Number]
		if !ok {
			continue
		}
		rows = append(rows, []any{
			asnID,
			record.Ipv6Prefixes,
			record.Ipv6PrefixesAgg,
			record.Ipv6LargestPrefix,
			record.Ipv6TotalAddresses,
		})
	}

	for start := 0; start < len(rows); start += metadataBatchSize {
		end := start + metadataBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[start:end]
		if len(batch) == 0 {
			continue
		}
		if err := database.BulkInsert(ctx, db, "ipv6_stat",
			[]string{"asn_id", "prefixes", "prefixes_aggregated", "largest_prefix", "total_addresses"},
			batch,
		); err != nil {
			return fmt.Errorf("insert ipv6_stat batch %d: %w", start, err)
		}
	}

	return nil
}
