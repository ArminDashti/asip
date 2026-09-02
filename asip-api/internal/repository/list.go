package repository

import (
	"context"
)

type CountryRow struct {
	Code string
	Name string
}

func (r *AsRepository) ListCountries(ctx context.Context) ([]CountryRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT country_code, country
FROM country
ORDER BY country_code
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []CountryRow
	for rows.Next() {
		var row CountryRow
		if err := rows.Scan(&row.Code, &row.Name); err != nil {
			return nil, err
		}
		countries = append(countries, row)
	}

	return countries, rows.Err()
}

type AsnRow struct {
	Number int
	Name   string
}

func (r *AsRepository) ListAsns(ctx context.Context) ([]AsnRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT asn, "as"
FROM asn
ORDER BY asn
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var asns []AsnRow
	for rows.Next() {
		var row AsnRow
		if err := rows.Scan(&row.Number, &row.Name); err != nil {
			return nil, err
		}
		asns = append(asns, row)
	}

	return asns, rows.Err()
}
