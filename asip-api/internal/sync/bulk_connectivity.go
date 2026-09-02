package sync

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

func insertConnectivity(
	ctx context.Context,
	db *sql.DB,
	relations []connectivityRelation,
	asnIDByNumber map[int]int,
) error {
	for start := 0; start < len(relations); start += metadataBatchSize {
		end := start + metadataBatchSize
		if end > len(relations) {
			end = len(relations)
		}
		batch := relations[start:end]
		rows := make([][]any, 0, len(batch))
		seen := make(map[string]struct{}, len(batch))
		for _, relation := range batch {
			asnID, ok := asnIDByNumber[relation.AsnNumber]
			if !ok {
				continue
			}
			providerID, ok := asnIDByNumber[relation.ProviderAsnNumber]
			if !ok || asnID == providerID {
				continue
			}
			key := fmt.Sprintf("%d:%d", asnID, providerID)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			rows = append(rows, []any{asnID, providerID, 0, 0, 0, nil, nil})
		}
		if len(rows) == 0 {
			continue
		}

		if err := database.BulkInsert(ctx, db, "connectivity",
			[]string{"asn_id", "provider_asn_id", "customers", "peers", "unclassified", "degree", "reach"},
			rows,
		); err != nil {
			return fmt.Errorf("insert connectivity batch %d: %w", start, err)
		}
	}

	return nil
}
