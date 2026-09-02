package sync

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

func insertAsns(
	ctx context.Context,
	db *sql.DB,
	records []asnRow,
	countryIDByCode map[string]int,
	originIDByName map[string]int,
	categoryIDByName map[string]int,
) (map[int]int, error) {
	for start := 0; start < len(records); start += metadataBatchSize {
		end := start + metadataBatchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[start:end]
		rows := make([][]any, 0, len(batch))
		for _, record := range batch {
			var countryID, originID, categoryID any
			if id, ok := countryIDByCode[record.CountryCode]; ok {
				countryID = id
			}
			if record.Origin != nil {
				if id, ok := originIDByName[*record.Origin]; ok {
					originID = id
				}
			}
			if record.Category != nil {
				if id, ok := categoryIDByName[*record.Category]; ok {
					categoryID = id
				}
			}
			rows = append(rows, []any{
				record.Number,
				record.Name,
				countryID,
				originID,
				categoryID,
				nil,
				record.Registered,
				record.LastModified,
				record.LastAnnounced,
				nil,
			})
		}

		if err := database.BulkInsert(ctx, db, "asn",
			[]string{"asn", "as", "country_id", "origin_id", "category_id", "network_role_id", "registered", "last_modified", "last_announced", "prefixes_last_modified"},
			rows,
		); err != nil {
			return nil, fmt.Errorf("insert asn batch %d: %w", start, err)
		}
	}

	idByNumber := make(map[int]int, len(records))
	queryRows, err := db.QueryContext(ctx, `SELECT id, asn FROM asn`)
	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var id int
		var number int
		if scanErr := queryRows.Scan(&id, &number); scanErr != nil {
			return nil, scanErr
		}
		idByNumber[number] = id
	}

	return idByNumber, queryRows.Err()
}
