package internals

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles all database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new store with optimized SQLite settings
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Connection pool optimization
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// SQLite optimizations
	pragmas := []string{
		"PRAGMA journal_mode=WAL",   // Write-Ahead Logging for concurrency
		"PRAGMA synchronous=NORMAL", // Faster writes, still safe
		"PRAGMA cache_size=-64000",  // 64MB cache
		"PRAGMA temp_store=MEMORY",  // In-memory temp tables
		"PRAGMA busy_timeout=5000",  // 5 second busy timeout
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, err
		}
	}

	// Create schema
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("database initialized with optimizations", "path", dbPath)
	return &Store{db: db}, nil
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS snippets (
		room_id TEXT NOT NULL,
		snippet_id TEXT PRIMARY KEY,
		ciphertext TEXT NOT NULL,
		iv TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_room_created ON snippets(room_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_created_at ON snippets(created_at);
	`
	_, err := db.Exec(schema)
	return err
}

// SaveSnippet stores a snippet in the database
func (s *Store) SaveSnippet(snippet *Snippet) error {
	_, err := s.db.Exec(
		"INSERT INTO snippets (room_id, snippet_id, ciphertext, iv, created_at) VALUES (?, ?, ?, ?, ?)",
		snippet.RoomID, snippet.SnippetID, snippet.Ciphertext, snippet.IV, snippet.CreatedAt,
	)
	return err
}

// GetRecentSnippets retrieves the most recent snippets for a room
func (s *Store) GetRecentSnippets(roomID string, limit int) ([]Snippet, error) {
	rows, err := s.db.Query(
		"SELECT snippet_id, ciphertext, iv, created_at FROM snippets WHERE room_id = ? ORDER BY created_at DESC LIMIT ?",
		roomID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []Snippet
	for rows.Next() {
		var snippet Snippet
		snippet.RoomID = roomID
		if err := rows.Scan(&snippet.SnippetID, &snippet.Ciphertext, &snippet.IV, &snippet.CreatedAt); err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse to send oldest first
	for i, j := 0, len(snippets)-1; i < j; i, j = i+1, j-1 {
		snippets[i], snippets[j] = snippets[j], snippets[i]
	}

	return snippets, nil
}

// CleanupOldSnippets removes snippets older than the specified TTL
func (s *Store) CleanupOldSnippets(ttl time.Duration) (int64, error) {
	cutoff := time.Now().Add(-ttl)
	result, err := s.db.Exec("DELETE FROM snippets WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Ping checks if the database connection is alive
func (s *Store) Ping() error {
	return s.db.Ping()
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
