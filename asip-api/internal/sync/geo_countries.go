package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadCountryIDMap(ctx context.Context, db *sql.DB) (map[string]int, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, country_code FROM country`)
	if err != nil {
		return nil, fmt.Errorf("load country map: %w", err)
	}
	defer rows.Close()

	idByCode := make(map[string]int)
	for rows.Next() {
		var id int
		var code string
		if scanErr := rows.Scan(&id, &code); scanErr != nil {
			return nil, scanErr
		}
		idByCode[strings.ToUpper(code)] = id
	}

	return idByCode, rows.Err()
}

func ensureGeoCountries(ctx context.Context, db *sql.DB, repoPath string) error {
	countryRoot := filepath.Join(repoPath, "country")
	countryDirs, err := os.ReadDir(countryRoot)
	if err != nil {
		return fmt.Errorf("read country directory: %w", err)
	}

	for _, countryDir := range countryDirs {
		if !countryDir.IsDir() {
			continue
		}

		countryCode := strings.ToUpper(countryDir.Name())
		jsonPath := filepath.Join(countryRoot, countryDir.Name(), countryCode+".json")
		if _, statErr := os.Stat(jsonPath); statErr != nil {
			jsonPath = filepath.Join(countryRoot, countryDir.Name(), strings.ToLower(countryCode)+".json")
		}

		content, readErr := os.ReadFile(jsonPath)
		if readErr != nil {
			continue
		}

		var geoEntry geoCountryEntry
		if unmarshalErr := json.Unmarshal(content, &geoEntry); unmarshalErr != nil {
			continue
		}

		name := geoEntry.Country
		if name == "" {
			name = countryCode
		}
		code := geoEntry.CountryCode
		if code == "" {
			code = countryCode
		}

		_, execErr := db.ExecContext(ctx, `
INSERT INTO country (country, country_code)
VALUES ($1, $2)
ON CONFLICT (country_code) DO UPDATE SET country = excluded.country
`, name, strings.ToUpper(code))
		if execErr != nil {
			return fmt.Errorf("upsert country %s: %w", code, execErr)
		}
	}

	return nil
}
