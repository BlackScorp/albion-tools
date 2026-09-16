// Package storage persists the latest market observations in SQLite.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackscorp/albion-helper/internal/catalog"
	_ "modernc.org/sqlite"
)

const (
	databaseFileName = "albion-helper.db"
	schemaVersion    = 1
)

// Repository stores the latest quote for every item, market and quality.
type Repository struct {
	db *sql.DB
}

// DefaultDatabasePath returns the per-user database location for the current OS.
func DefaultDatabasePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}
	return filepath.Join(configDir, "Albion Helper", databaseFileName), nil
}

// Open creates or opens a database and applies all required migrations.
func Open(path string) (*Repository, error) {
	if path == "" {
		return nil, fmt.Errorf("database path must not be empty")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	// In-memory databases are connection-local in SQLite. A single connection
	// also keeps repository operations predictable for the small desktop app.
	db.SetMaxOpenConns(1)
	repository := &Repository{db: db}
	if err := repository.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return repository, nil
}

// Close releases the database connection.
func (r *Repository) Close() error { return r.db.Close() }

func (r *Repository) migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS prices (
			item_id TEXT NOT NULL,
			market TEXT NOT NULL,
			quality INTEGER NOT NULL,
			buy INTEGER NOT NULL,
			sell INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (item_id, market, quality)
		);
	`)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, unixepoch()) ON CONFLICT(version) DO NOTHING`,
		schemaVersion,
	)
	if err != nil {
		return fmt.Errorf("record schema migration: %w", err)
	}
	return nil
}

// UpsertPrices atomically replaces the latest values for the supplied quotes.
func (r *Repository) UpsertPrices(ctx context.Context, prices []catalog.Price) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin price update: %w", err)
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO prices (item_id, market, quality, buy, sell, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(item_id, market, quality) DO UPDATE SET
			buy = excluded.buy,
			sell = excluded.sell,
			updated_at = excluded.updated_at`
	for _, price := range prices {
		if price.ItemID == "" || price.Market == "" {
			return fmt.Errorf("invalid price key: item=%q market=%q", price.ItemID, price.Market)
		}
		if _, err := tx.ExecContext(ctx, query, price.ItemID, price.Market, price.Quality, price.Buy, price.Sell, price.UpdatedAt); err != nil {
			return fmt.Errorf("upsert price %s/%s: %w", price.ItemID, price.Market, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit price update: %w", err)
	}
	return nil
}

// Prices returns all stored observations in stable key order.
func (r *Repository) Prices(ctx context.Context) ([]catalog.Price, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT item_id, market, quality, buy, sell, updated_at
		FROM prices ORDER BY item_id, market, quality`)
	if err != nil {
		return nil, fmt.Errorf("read prices: %w", err)
	}
	defer rows.Close()

	var prices []catalog.Price
	for rows.Next() {
		var price catalog.Price
		if err := rows.Scan(&price.ItemID, &price.Market, &price.Quality, &price.Buy, &price.Sell, &price.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, price)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read prices: %w", err)
	}
	return prices, nil
}
