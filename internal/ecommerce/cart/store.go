package cart

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists carts to a SQLite database.
type SQLiteStore struct {
	db   *sql.DB
	done chan struct{}
}

// NewSQLiteStore opens (or creates) the SQLite database at dbPath and
// ensures the schema exists.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	const schema = `CREATE TABLE IF NOT EXISTS cart_items (
		session_id     TEXT NOT NULL,
		product_id     TEXT NOT NULL,
		name           TEXT NOT NULL,
		price_cents    INTEGER NOT NULL,
		quantity       INTEGER NOT NULL,
		updated_at     INTEGER NOT NULL DEFAULT 0,
		stock_limit    INTEGER NOT NULL DEFAULT -1,
		variable_price INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (session_id, product_id)
	)`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	// Add columns to existing tables (idempotent).
	// DEFAULT must be a constant for ALTER TABLE ADD COLUMN in SQLite.
	db.Exec("ALTER TABLE cart_items ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE cart_items ADD COLUMN stock_limit INTEGER NOT NULL DEFAULT -1")
	db.Exec("ALTER TABLE cart_items ADD COLUMN variable_price INTEGER NOT NULL DEFAULT 0")

	store := &SQLiteStore{db: db, done: make(chan struct{})}
	go store.cleanupLoop()
	return store, nil
}

// Get returns the cart for a session. If no rows exist an empty cart is returned.
func (s *SQLiteStore) Get(sessionID string) *Cart {
	c := &Cart{
		ID:    sessionID,
		Items: []CartItem{},
	}

	rows, err := s.db.Query(
		"SELECT product_id, name, price_cents, quantity, stock_limit, variable_price FROM cart_items WHERE session_id = ?",
		sessionID,
	)
	if err != nil {
		log.Printf("sqlite Get: %v", err)
		return c
	}
	defer rows.Close()

	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ProductID, &item.Name, &item.PriceCents, &item.Quantity, &item.StockLimit, &item.VariablePrice); err != nil {
			log.Printf("sqlite scan: %v", err)
			continue
		}
		item.Price = float64(item.PriceCents) / 100.0
		c.Items = append(c.Items, item)
	}

	return c
}

// Save persists the cart, replacing all existing rows for the session.
func (s *SQLiteStore) Save(sessionID string, c *Cart) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM cart_items WHERE session_id = ?", sessionID); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete: %w", err)
	}

	now := time.Now().Unix()
	stmt, err := tx.Prepare(
		"INSERT INTO cart_items (session_id, product_id, name, price_cents, quantity, updated_at, stock_limit, variable_price) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, item := range c.Items {
		if _, err := stmt.Exec(sessionID, item.ProductID, item.Name, item.PriceCents, item.Quantity, now, item.StockLimit, item.VariablePrice); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert %s: %w", item.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Delete removes all cart items for a session.
func (s *SQLiteStore) Delete(sessionID string) {
	if _, err := s.db.Exec("DELETE FROM cart_items WHERE session_id = ?", sessionID); err != nil {
		log.Printf("sqlite Delete: %v", err)
	}
}

// CountItems returns the total quantity of items in a session's cart.
func (s *SQLiteStore) CountItems(sessionID string) int {
	var count int
	err := s.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE session_id = ?",
		sessionID,
	).Scan(&count)
	if err != nil {
		log.Printf("sqlite CountItems: %v", err)
		return 0
	}
	return count
}

// cleanupLoop periodically removes cart items older than 7 days.
func (s *SQLiteStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.done:
			return
		}
	}
}

func (s *SQLiteStore) cleanup() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour).Unix()
	result, err := s.db.Exec("DELETE FROM cart_items WHERE updated_at < ?", cutoff)
	if err != nil {
		log.Printf("cart cleanup: %v", err)
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		log.Printf("cart cleanup: removed %d stale items", n)
	}
}

// Close stops the cleanup goroutine and closes the database.
func (s *SQLiteStore) Close() error {
	close(s.done)
	return s.db.Close()
}
