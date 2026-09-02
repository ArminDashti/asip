package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

const prefixBatchSize = 10000

func importAsPrefixes(ctx context.Context, db *sql.DB, repoPath string) error {
	asnIDByNumber, err := loadAsnIDMap(ctx, db)
	if err != nil {
		return err
	}

	asRoot := filepath.Join(repoPath, "as")
	entries, err := os.ReadDir(asRoot)
	if err != nil {
		return fmt.Errorf("read as directory: %w", err)
	}

	ipv4Batch := make([][]any, 0, prefixBatchSize)
	ipv6Batch := make([][]any, 0, prefixBatchSize)
	importedV4 := int64(0)
	importedV6 := int64(0)
	skipped := 0

	flushV4 := func() error {
		if len(ipv4Batch) == 0 {
			return nil
		}
		if err := database.BulkInsert(ctx, db, "asn_ipv4",
			[]string{"asn_id", "cidr", "last_modified", "start_ip", "end_ip"},
			ipv4Batch,
		); err != nil {
			return fmt.Errorf("insert asn_ipv4: %w", err)
		}
		importedV4 += int64(len(ipv4Batch))
		ipv4Batch = ipv4Batch[:0]
		return nil
	}

	flushV6 := func() error {
		if len(ipv6Batch) == 0 {
			return nil
		}
		if err := database.BulkInsert(ctx, db, "asn_ipv6",
			[]string{"asn_id", "cidr", "last_modified"},
			ipv6Batch,
		); err != nil {
			return fmt.Errorf("insert asn_ipv6: %w", err)
		}
		importedV6 += int64(len(ipv6Batch))
		ipv6Batch = ipv6Batch[:0]
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		asnNumber, parseErr := strconv.Atoi(entry.Name())
		if parseErr != nil {
			log.Printf("sync: skipping non-numeric as directory %q", entry.Name())
			continue
		}

		asnID, ok := asnIDByNumber[asnNumber]
		if !ok {
			skipped++
			continue
		}

		filePath := filepath.Join(asRoot, entry.Name(), "aggregated.json")
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			log.Printf("sync: skipping as%d: read error: %v", asnNumber, readErr)
			continue
		}

		var prefixEntry asPrefixEntry
		if unmarshalErr := json.Unmarshal(content, &prefixEntry); unmarshalErr != nil {
			log.Printf("sync: skipping as%d: unmarshal error: %v", asnNumber, unmarshalErr)
			continue
		}

		for _, cidr := range prefixEntry.Prefixes.Ipv4 {
			start, end, rangeErr := database.CIDRRange(cidr)
			if rangeErr != nil {
				log.Printf("sync: skipping prefix %q for as%d: %v", cidr, asnNumber, rangeErr)
				continue
			}
			ipv4Batch = append(ipv4Batch, []any{asnID, cidr, nil, start, end})
			if len(ipv4Batch) >= prefixBatchSize {
				if flushErr := flushV4(); flushErr != nil {
					return flushErr
				}
			}
		}

		for _, cidr := range prefixEntry.Prefixes.Ipv6 {
			ipv6Batch = append(ipv6Batch, []any{asnID, cidr, nil})
			if len(ipv6Batch) >= prefixBatchSize {
				if flushErr := flushV6(); flushErr != nil {
					return flushErr
				}
			}
		}
	}

	if err := flushV4(); err != nil {
		return err
	}
	if err := flushV6(); err != nil {
		return err
	}

	log.Printf("sync: imported %d ipv4 and %d ipv6 asn prefixes (%d as directories skipped — no asn row)",
		importedV4, importedV6, skipped)
	return nil
}

func loadAsnIDMap(ctx context.Context, db *sql.DB) (map[int]int, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, asn FROM asn`)
	if err != nil {
		return nil, fmt.Errorf("load asn map: %w", err)
	}
	defer rows.Close()

	idByNumber := make(map[int]int)
	for rows.Next() {
		var id int
		var number int
		if scanErr := rows.Scan(&id, &number); scanErr != nil {
			return nil, scanErr
		}
		idByNumber[number] = id
	}

	return idByNumber, rows.Err()
}
