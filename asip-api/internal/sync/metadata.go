package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const metadataBatchSize = 5000

func importMetadata(ctx context.Context, db *sql.DB, repoPath string) error {
	filePath := filepath.Join(repoPath, "as.json")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open as.json: %w", err)
	}
	defer file.Close()

	countries := make(map[string]countryRow)
	origins := make(map[string]string)
	categories := make(map[string]string)
	asnRows := make([]asnRow, 0, 700_000)
	connections := make([]connectivityRelation, 0, 2_000_000)

	decoder := json.NewDecoder(file)
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("read metadata json start: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("metadata json must be an array")
	}

	for decoder.More() {
		var entry metadataEntry
		if decodeErr := decoder.Decode(&entry); decodeErr != nil {
			return fmt.Errorf("decode metadata entry: %w", decodeErr)
		}

		meta := entry.Metadata
		if meta.Description == "" {
			if meta.Handle != "" {
				meta.Description = meta.Handle
			} else {
				meta.Description = fmt.Sprintf("AS%d", entry.Asn)
			}
		}
		if meta.CountryCode == "" {
			meta.CountryCode = "XX"
		}
		if meta.Country == "" {
			meta.Country = meta.CountryCode
		}

		countries[meta.CountryCode] = countryRow{Code: meta.CountryCode, Name: meta.Country}
		if meta.Origin != nil && *meta.Origin != "" {
			origins[*meta.Origin] = *meta.Origin
		}
		if meta.Category != nil && *meta.Category != "" {
			categories[*meta.Category] = *meta.Category
		}

		asnRecord := asnRow{
			Number:        entry.Asn,
			Name:          meta.Description,
			Category:      meta.Category,
			Origin:        meta.Origin,
			Registered:    parseDate(meta.Registered),
			LastModified:  parseDate(meta.LastModified),
			LastAnnounced: parseDate(entry.LastAnnounced),
			CountryCode:   meta.CountryCode,
		}

		if entry.Stats != nil && entry.Stats.Ipv4 != nil {
			asnRecord.Ipv4Prefixes = entry.Stats.Ipv4.Prefixes
			asnRecord.Ipv4PrefixesAgg = entry.Stats.Ipv4.PrefixesAggregated
			asnRecord.Ipv4LargestPrefix = entry.Stats.Ipv4.LargestPrefix
			if entry.Stats.Ipv4.TotalAddresses != nil {
				asnRecord.Ipv4TotalAddresses = *entry.Stats.Ipv4.TotalAddresses
			}
		}
		if entry.Stats != nil && entry.Stats.Ipv6 != nil {
			asnRecord.Ipv6Prefixes = entry.Stats.Ipv6.Prefixes
			asnRecord.Ipv6PrefixesAgg = entry.Stats.Ipv6.PrefixesAggregated
			asnRecord.Ipv6LargestPrefix = entry.Stats.Ipv6.LargestPrefix
			if entry.Stats.Ipv6.TotalAddresses != nil {
				asnRecord.Ipv6TotalAddresses = *entry.Stats.Ipv6.TotalAddresses
			}
		}
		asnRows = append(asnRows, asnRecord)

		if entry.Stats != nil && entry.Stats.Connectivity != nil {
			for _, providerAsn := range entry.Stats.Connectivity.ProviderAsns {
				connections = append(connections, connectivityRelation{
					AsnNumber:         entry.Asn,
					ProviderAsnNumber: providerAsn,
				})
			}
		}
	}

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("read metadata json end: %w", err)
	}

	log.Printf("sync: metadata parsed — %d countries, %d as records, %d connectivity relations",
		len(countries), len(asnRows), len(connections))

	countryIDByCode, err := insertCountries(ctx, db, countries)
	if err != nil {
		return err
	}

	originIDByName, err := insertLookupValues(ctx, db, "origin", "origin", origins)
	if err != nil {
		return err
	}

	categoryIDByName, err := insertLookupValues(ctx, db, "category", "category", categories)
	if err != nil {
		return err
	}

	asnIDByNumber, err := insertAsns(ctx, db, asnRows, countryIDByCode, originIDByName, categoryIDByName)
	if err != nil {
		return err
	}

	if err := insertIpv4Stats(ctx, db, asnRows, asnIDByNumber); err != nil {
		return err
	}

	if err := insertIpv6Stats(ctx, db, asnRows, asnIDByNumber); err != nil {
		return err
	}

	if err := insertConnectivity(ctx, db, connections, asnIDByNumber); err != nil {
		return err
	}

	return nil
}
