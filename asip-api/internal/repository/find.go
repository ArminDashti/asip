package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/database"
	"github.com/ArminDashti/as-ip/server/internal/model"
)

func (r *AsRepository) FindCountryByIP(ctx context.Context, ip string) (countryCode, countryName string, err error) {
	ipInt, err := database.IPv4ToInt(ip)
	if err != nil {
		return "", "", fmt.Errorf("invalid ip: %w", err)
	}

	query := `
SELECT c.country_code, c.country
FROM country_ipv4 p
JOIN country c ON c.id = p.country_id
WHERE $1 >= p.start_ip AND $2 <= p.end_ip
ORDER BY (p.end_ip - p.start_ip) ASC
LIMIT 1
`

	var code, name sql.NullString
	scanErr := r.db.QueryRowContext(ctx, query, ipInt, ipInt).Scan(&code, &name)
	if scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return "", "", ErrNotFound
		}
		return "", "", scanErr
	}

	if code.Valid {
		countryCode = code.String
	}
	if name.Valid {
		countryName = name.String
	}
	return countryCode, countryName, nil
}

func (r *AsRepository) FindByIP(ctx context.Context, ip string) (model.AsRecord, error) {
	ipInt, err := database.IPv4ToInt(ip)
	if err != nil {
		return model.AsRecord{}, fmt.Errorf("invalid ip: %w", err)
	}

	query := asRecordSelect + `
WHERE EXISTS (
    SELECT 1
    FROM asn_ipv4 p
    WHERE p.asn_id = a.id
      AND $1 >= p.start_ip AND $2 <= p.end_ip
)
ORDER BY s.largest_prefix IS NULL, s.largest_prefix DESC, a.asn ASC
LIMIT 1
`

	return r.scanSingle(ctx, query, ipInt, ipInt)
}

func (r *AsRepository) FindByAsnNumber(ctx context.Context, asnNumber int) (model.AsRecord, error) {
	query := asRecordSelect + `WHERE a.asn = $1 LIMIT 1`
	return r.scanSingle(ctx, query, asnNumber)
}

func (r *AsRepository) FindByAsIdentifier(ctx context.Context, asIdentifier string) (model.AsRecord, error) {
	normalized := strings.TrimSpace(asIdentifier)
	if normalized == "" {
		return model.AsRecord{}, fmt.Errorf("as identifier is required")
	}

	if asnNumber, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(normalized), "AS")); err == nil {
		record, lookupErr := r.FindByAsnNumber(ctx, asnNumber)
		if lookupErr == nil {
			return record, nil
		}
		if !errors.Is(lookupErr, ErrNotFound) {
			return model.AsRecord{}, lookupErr
		}
	}

	query := asRecordSelect + `
WHERE LOWER(a."as") LIKE LOWER('%' || $1 || '%')
ORDER BY a.asn ASC
LIMIT 1
`
	return r.scanSingle(ctx, query, normalized)
}

func (r *AsRepository) FindByAsnIdentifier(ctx context.Context, asnIdentifier string) (model.AsRecord, error) {
	normalized := strings.TrimSpace(asnIdentifier)
	if normalized == "" {
		return model.AsRecord{}, fmt.Errorf("asn identifier is required")
	}

	if asnNumber, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(normalized), "AS")); err == nil {
		return r.FindByAsnNumber(ctx, asnNumber)
	}

	query := asRecordSelect + `
WHERE LOWER(a."as") LIKE LOWER('%' || $1 || '%')
ORDER BY a.asn ASC
LIMIT 1
`
	return r.scanSingle(ctx, query, normalized)
}

func (r *AsRepository) ListByCountry(ctx context.Context, countryCode string) ([]model.AsRecord, error) {
	query := asRecordSelect + `
WHERE LOWER(c.country_code) = LOWER($1)
ORDER BY a.asn ASC
`
	return r.scanMany(ctx, query, countryCode)
}
