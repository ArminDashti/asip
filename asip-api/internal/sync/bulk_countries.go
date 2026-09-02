package sync

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

func insertCountries(ctx context.Context, db *sql.DB, countries map[string]countryRow) (map[string]int, error) {
	rows := make([][]any, 0, len(countries))
	for _, country := range countries {
		rows = append(rows, []any{country.Name, country.Code})
	}

	if err := database.BulkInsert(ctx, db, "country", []string{"country", "country_code"}, rows); err != nil {
		return nil, fmt.Errorf("insert countries: %w", err)
	}

	idByCode := make(map[string]int, len(countries))
	queryRows, err := db.QueryContext(ctx, `SELECT id, country_code FROM country`)
	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var id int
		var code string
		if scanErr := queryRows.Scan(&id, &code); scanErr != nil {
			return nil, scanErr
		}
		idByCode[code] = id
	}

	return idByCode, queryRows.Err()
}

func insertLookupValues(
	ctx context.Context,
	db *sql.DB,
	table string,
	valueColumn string,
	values map[string]string,
) (map[string]int, error) {
	if len(values) == 0 {
		return map[string]int{}, nil
	}

	rows := make([][]any, 0, len(values))
	for _, value := range values {
		rows = append(rows, []any{value, value})
	}

	columns := []string{valueColumn, "description"}
	if err := database.BulkInsert(ctx, db, table, columns, rows); err != nil {
		return nil, fmt.Errorf("insert %s: %w", table, err)
	}

	idByValue := make(map[string]int, len(values))
	query := fmt.Sprintf("SELECT id, %s FROM %s", valueColumn, table)
	queryRows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var id int
		var value string
		if scanErr := queryRows.Scan(&id, &value); scanErr != nil {
			return nil, scanErr
		}
		idByValue[value] = id
	}

	return idByValue, queryRows.Err()
}
