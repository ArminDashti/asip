package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/model"
)

const asRecordSelect = `
SELECT
    a.id,
    a."as",
    cat.category,
    a.registered,
    orig.origin,
    a.last_modified,
    a.last_announced,
    c.country_code,
    c.country,
    a.asn,
    COALESCE(s.prefixes, 0),
    COALESCE(s.prefixes_aggregated, 0),
    s.largest_prefix,
    COALESCE(s.total_addresses, 0),
    COALESCE((
        SELECT STRING_AGG(provider.asn::text, ',' ORDER BY provider.asn)
        FROM connectivity conn
        JOIN asn provider ON provider.id = conn.provider_asn_id
        WHERE conn.asn_id = a.id
    ), '') AS provider_asns
FROM asn a
LEFT JOIN country c ON c.id = a.country_id
LEFT JOIN category cat ON cat.id = a.category_id
LEFT JOIN origin orig ON orig.id = a.origin_id
LEFT JOIN ipv4_stat s ON s.asn_id = a.id
`

type scannable interface {
	Scan(dest ...any) error
}

func (r *AsRepository) scanSingle(ctx context.Context, query string, args ...any) (model.AsRecord, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	record, err := scanAsRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AsRecord{}, ErrNotFound
		}
		return model.AsRecord{}, err
	}
	return record, nil
}

func (r *AsRepository) scanMany(ctx context.Context, query string, args ...any) ([]model.AsRecord, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.AsRecord
	for rows.Next() {
		record, scanErr := scanAsRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func scanAsRecord(row scannable) (model.AsRecord, error) {
	var record model.AsRecord
	var registered sql.NullString
	var lastModified sql.NullString
	var lastAnnounced sql.NullString
	var largestPrefix sql.NullInt64
	var providerAsns string
	err := row.Scan(
		&record.ID,
		&record.Name,
		&record.Category,
		&registered,
		&record.Origin,
		&lastModified,
		&lastAnnounced,
		&record.CountryCode,
		&record.CountryName,
		&record.AsnNumber,
		&record.PrefixCount,
		&record.PrefixesAgg,
		&largestPrefix,
		&record.TotalAddresses,
		&providerAsns,
	)
	if err != nil {
		return model.AsRecord{}, err
	}
	record.Registered = parseStoredTime(registered)
	record.LastModified = parseStoredTime(lastModified)
	record.LastAnnounced = parseStoredTime(lastAnnounced)
	if largestPrefix.Valid {
		value := int(largestPrefix.Int64)
		record.LargestPrefix = &value
	}
	record.ProviderAsns = parseIntList(providerAsns)
	return record, nil
}

func parseStoredTime(value sql.NullString) *time.Time {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 +0000 UTC",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value.String); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseIntList(value string) []int {
	if value == "" {
		return []int{}
	}
	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			result = append(result, n)
		}
	}
	return result
}
