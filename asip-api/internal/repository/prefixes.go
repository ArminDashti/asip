package repository

import (
	"context"
)

func (r *AsRepository) LoadPrefixes(ctx context.Context, asnID int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT cidr FROM asn_ipv4 WHERE asn_id = $1 ORDER BY cidr
`, asnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []string
	for rows.Next() {
		var prefix string
		if err := rows.Scan(&prefix); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, rows.Err()
}
