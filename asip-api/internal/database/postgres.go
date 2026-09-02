package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, dbConfig config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS sync_state (
		id           INTEGER PRIMARY KEY CHECK (id = 1),
		last_sync_at TIMESTAMPTZ
	)`)
	if err != nil {
		return fmt.Errorf("migrate sync_state: %w", err)
	}
	return nil
}

func initSchema(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'country'
		)
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if exists {
		return nil
	}

	schemaPath := os.Getenv("DB_SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "db/init.sql"
	}

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaPath, err)
	}

	statements := strings.Split(string(content), ";")
	for _, statement := range statements {
		stmt := strings.TrimSpace(statement)
		if stmt == "" {
			continue
		}
		if _, execErr := db.Exec(stmt); execErr != nil {
			return fmt.Errorf("apply schema: %w\nstatement: %s", execErr, stmt)
		}
	}

	return nil
}
