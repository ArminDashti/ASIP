package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/config"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, dbConfig config.DBConfig) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbConfig.Path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", filepath.ToSlash(dbConfig.Path))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := initSchema(db, dbConfig.SchemaPath); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func initSchema(db *sql.DB, schemaPath string) error {
	var count int
	err := db.QueryRow(`
SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'country'
`).Scan(&count)
	if err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if count > 0 {
		return nil
	}

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