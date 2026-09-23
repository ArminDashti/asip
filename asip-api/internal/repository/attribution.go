package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/model"
)

// FindAttribution returns the latest stored AS/country attribution for ip.
// Returns ErrNotFound when the IP was never resolved.
func (r *AsRepository) FindAttribution(ctx context.Context, ip string) (model.IpAttribution, error) {
	var attribution model.IpAttribution
	var asn sql.NullInt64
	var asName, countryCode, country, source sql.NullString
	var resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
SELECT ip, asn, as_name, country_code, country, source, resolved_at
FROM ip_attribution
WHERE ip = $1
`, ip).Scan(&attribution.Ip, &asn, &asName, &countryCode, &country, &source, &resolvedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IpAttribution{}, ErrNotFound
		}
		return model.IpAttribution{}, err
	}

	if asn.Valid {
		attribution.Asn = int(asn.Int64)
	}
	attribution.AsName = asName.String
	attribution.CountryCode = countryCode.String
	attribution.Country = country.String
	attribution.Source = source.String
	if resolvedAt.Valid {
		attribution.ResolvedAt = resolvedAt.Time
	}
	return attribution, nil
}

// UpsertAttribution stores the latest attribution for an IP, overwriting the
// previous value in place (no archive rows are kept).
func (r *AsRepository) UpsertAttribution(ctx context.Context, attribution model.IpAttribution) error {
	resolvedAt := attribution.ResolvedAt
	if resolvedAt.IsZero() {
		resolvedAt = time.Now().UTC()
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO ip_attribution (ip, asn, as_name, country_code, country, source, resolved_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (ip) DO UPDATE SET
    asn          = EXCLUDED.asn,
    as_name      = EXCLUDED.as_name,
    country_code = EXCLUDED.country_code,
    country      = EXCLUDED.country,
    source       = EXCLUDED.source,
    resolved_at  = EXCLUDED.resolved_at
`, attribution.Ip, attribution.Asn, attribution.AsName, attribution.CountryCode,
		attribution.Country, attribution.Source, resolvedAt)
	if err != nil {
		return fmt.Errorf("upsert ip attribution: %w", err)
	}
	return nil
}

// FindAsnName returns the handle ("as") of an ASN when the imported dataset
// knows it. Used to name an ASN that was discovered outside the dataset.
func (r *AsRepository) FindAsnName(ctx context.Context, asn int) (string, error) {
	var name sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT "as" FROM asn WHERE asn = $1 LIMIT 1`, asn).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return strings.TrimSpace(name.String), nil
}

// FindCountryNameByCode resolves an ISO 3166-1 alpha-2 code to a display name.
func (r *AsRepository) FindCountryNameByCode(ctx context.Context, code string) (string, error) {
	normalized := strings.TrimSpace(code)
	if normalized == "" {
		return "", ErrNotFound
	}

	var name sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT country FROM country WHERE LOWER(country_code) = LOWER($1) LIMIT 1
`, normalized).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return strings.TrimSpace(name.String), nil
}
