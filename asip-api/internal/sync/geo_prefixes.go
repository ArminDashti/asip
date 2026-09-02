package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/database"
)

const geoBatchSize = 10000

func importGeoPrefixes(ctx context.Context, db *sql.DB, repoPath string) error {
	countryIDByCode, err := loadCountryIDMap(ctx, db)
	if err != nil {
		return err
	}

	countryRoot := filepath.Join(repoPath, "country")
	countryDirs, err := os.ReadDir(countryRoot)
	if err != nil {
		return fmt.Errorf("read country directory: %w", err)
	}

	ipv4Batch := make([][]any, 0, geoBatchSize)
	ipv6Batch := make([][]any, 0, geoBatchSize)
	importedV4 := int64(0)
	importedV6 := int64(0)

	flushV4 := func() error {
		if len(ipv4Batch) == 0 {
			return nil
		}
		if err := database.BulkInsert(ctx, db, "country_ipv4",
			[]string{"country_id", "cidr", "last_modified", "start_ip", "end_ip"},
			ipv4Batch,
		); err != nil {
			return fmt.Errorf("insert country_ipv4: %w", err)
		}
		importedV4 += int64(len(ipv4Batch))
		ipv4Batch = ipv4Batch[:0]
		return nil
	}

	flushV6 := func() error {
		if len(ipv6Batch) == 0 {
			return nil
		}
		if err := database.BulkInsert(ctx, db, "country_ipv6",
			[]string{"country_id", "cidr", "last_modified"},
			ipv6Batch,
		); err != nil {
			return fmt.Errorf("insert country_ipv6: %w", err)
		}
		importedV6 += int64(len(ipv6Batch))
		ipv6Batch = ipv6Batch[:0]
		return nil
	}

	appendPrefixes := func(countryID int, family string, prefixes []string) error {
		for _, cidr := range prefixes {
			if family == "ipv4" {
				start, end, rangeErr := database.CIDRRange(cidr)
				if rangeErr != nil {
					log.Printf("sync: skipping geo prefix %q: %v", cidr, rangeErr)
					continue
				}
				ipv4Batch = append(ipv4Batch, []any{countryID, cidr, nil, start, end})
				if len(ipv4Batch) >= geoBatchSize {
					if flushErr := flushV4(); flushErr != nil {
						return flushErr
					}
				}
			} else {
				ipv6Batch = append(ipv6Batch, []any{countryID, cidr, nil})
				if len(ipv6Batch) >= geoBatchSize {
					if flushErr := flushV6(); flushErr != nil {
						return flushErr
					}
				}
			}
		}
		return nil
	}

	for _, countryDir := range countryDirs {
		if !countryDir.IsDir() {
			continue
		}

		countryCode := strings.ToUpper(countryDir.Name())
		countryID, ok := countryIDByCode[countryCode]
		if !ok {
			continue
		}

		jsonPath := filepath.Join(countryRoot, countryDir.Name(), countryCode+".json")
		if _, statErr := os.Stat(jsonPath); statErr != nil {
			jsonPath = filepath.Join(countryRoot, countryDir.Name(), strings.ToLower(countryCode)+".json")
			if _, statErr2 := os.Stat(jsonPath); statErr2 != nil {
				log.Printf("sync: skipping country %s: no data file found", countryCode)
				continue
			}
		}

		content, readErr := os.ReadFile(jsonPath)
		if readErr != nil {
			log.Printf("sync: skipping country %s: read error: %v", countryCode, readErr)
			continue
		}

		var geoEntry geoCountryEntry
		if unmarshalErr := json.Unmarshal(content, &geoEntry); unmarshalErr != nil {
			log.Printf("sync: skipping country %s: unmarshal error: %v", countryCode, unmarshalErr)
			continue
		}

		if appendErr := appendPrefixes(countryID, "ipv4", geoEntry.Prefixes.Ipv4); appendErr != nil {
			return appendErr
		}
		if appendErr := appendPrefixes(countryID, "ipv6", geoEntry.Prefixes.Ipv6); appendErr != nil {
			return appendErr
		}
	}

	if err := flushV4(); err != nil {
		return err
	}
	if err := flushV6(); err != nil {
		return err
	}

	log.Printf("sync: imported %d country ipv4 and %d country ipv6 prefixes", importedV4, importedV6)
	return nil
}
