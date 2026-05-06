// Package database provides SQLite persistence for users, sessions, and submissions.
// Tables are auto-created on startup if they don't exist (zero-config migrations).
package database

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// sessionSecret is the HMAC key for signing session tokens.
// Loaded from SESSION_SECRET env var. If empty, tokens are unsigned (dev mode).
var sessionSecret string

func init() {
	sessionSecret = os.Getenv("SESSION_SECRET")
}

// DB wraps the sql.DB connection and provides repository methods.
type DB struct {
	conn *sql.DB
}

// User represents a registered user (authenticated via GitHub OAuth).
type User struct {
	ID        int64  `json:"id"`
	GitHubID  int64  `json:"github_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
	Role      string `json:"role"` // "user" or "admin"
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Session represents an active user session.
type Session struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Submission represents a user's code/repo submission and its evaluation result.
type Submission struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Type      string `json:"type"`    // "file" or "repo"
	Name      string `json:"name"`    // filename or repo URL
	Score     int    `json:"score"`   // 0-100
	CID       string `json:"cid"`     // IPFS CID of the evaluation report
	TxHash    string `json:"tx_hash"` // Flow transaction hash
	CreatedAt string `json:"created_at"`
}

// Open connects to the SQLite database at the given path and runs migrations.
// If DATABASE_URL env var is set, it overrides the path argument.
// Creates the directory and file if they don't exist.
func Open(path string) (*DB, error) {
	if envPath := os.Getenv("DATABASE_URL"); envPath != "" {
		path = envPath
	}
	if path == "" {
		path = "./data/posa.db"
	}

	// Ensure the directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("database: failed to create directory %s: %w", dir, err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open %s: %w", path, err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("database: failed to set WAL mode: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate creates tables if they don't exist.
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		github_id INTEGER UNIQUE NOT NULL,
		username TEXT NOT NULL,
		avatar_url TEXT DEFAULT '',
		email TEXT DEFAULT '',
		role TEXT DEFAULT 'user',
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		token_hash TEXT UNIQUE NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS submissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		score INTEGER DEFAULT 0,
		cid TEXT DEFAULT '',
		tx_hash TEXT DEFAULT '',
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_hash);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_submissions_user ON submissions(user_id);
	CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id);
	`
	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("database: migration failed: %w", err)
	}
	return nil
}

// --- User Repository ---

// CreateUser inserts a new user or updates an existing one (upsert on github_id).
func (db *DB) CreateUser(githubID int64, username, avatarURL, email string) (*User, error) {
	_, err := db.conn.Exec(`
		INSERT INTO users (github_id, username, avatar_url, email)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(github_id) DO UPDATE SET
			username = excluded.username,
			avatar_url = excluded.avatar_url,
			email = excluded.email,
			updated_at = datetime('now')
	`, githubID, username, avatarURL, email)
	if err != nil {
		return nil, fmt.Errorf("database: create user failed: %w", err)
	}
	return db.GetUserByGitHubID(githubID)
}

// GetUserByID retrieves a user by their internal ID.
func (db *DB) GetUserByID(id int64) (*User, error) {
	var u User
	err := db.conn.QueryRow(`
		SELECT id, github_id, username, avatar_url, email, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&u.ID, &u.GitHubID, &u.Username, &u.AvatarURL, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database: get user failed: %w", err)
	}
	return &u, nil
}

// GetUserByGitHubID retrieves a user by their GitHub ID.
func (db *DB) GetUserByGitHubID(githubID int64) (*User, error) {
	var u User
	err := db.conn.QueryRow(`
		SELECT id, github_id, username, avatar_url, email, role, created_at, updated_at
		FROM users WHERE github_id = ?
	`, githubID).Scan(&u.ID, &u.GitHubID, &u.Username, &u.AvatarURL, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database: get user by github_id failed: %w", err)
	}
	return &u, nil
}

// ListUsers returns all users with pagination.
func (db *DB) ListUsers(limit, offset int) ([]*User, error) {
	rows, err := db.conn.Query(`
		SELECT id, github_id, username, avatar_url, email, role, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database: list users failed: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.GitHubID, &u.Username, &u.AvatarURL, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// UpdateUserRole sets a user's role (e.g. "admin" or "user").
func (db *DB) UpdateUserRole(id int64, role string) error {
	_, err := db.conn.Exec(`UPDATE users SET role = ?, updated_at = datetime('now') WHERE id = ?`, role, id)
	return err
}

// CountUsers returns the total number of registered users.
func (db *DB) CountUsers() (int64, error) {
	var count int64
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// --- Session Repository ---

// CreateSession generates a new session token for a user and stores its hash.
// Returns the raw token (to be sent as a cookie).
func (db *DB) CreateSession(userID int64, duration time.Duration) (string, error) {
	token := generateSessionToken()
	hash := hashToken(token)
	expiresAt := time.Now().Add(duration).UTC().Format(time.RFC3339)

	_, err := db.conn.Exec(`
		INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)
	`, userID, hash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("database: create session failed: %w", err)
	}
	return token, nil
}

// GetSession validates a session token's HMAC signature, then looks up the
// associated user. Returns nil if the token is invalid, tampered, or expired.
func (db *DB) GetSession(token string) (*User, error) {
	if !ValidateToken(token) {
		return nil, nil
	}

	hash := hashToken(token)

	var userID int64
	var expiresAt string
	err := db.conn.QueryRow(`
		SELECT user_id, expires_at FROM sessions WHERE token_hash = ?
	`, hash).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database: get session failed: %w", err)
	}

	// Check expiry.
	expiry, _ := time.Parse(time.RFC3339, expiresAt)
	if time.Now().After(expiry) {
		db.conn.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hash)
		return nil, nil
	}

	return db.GetUserByID(userID)
}

// DeleteSession removes a session by its raw token.
func (db *DB) DeleteSession(token string) error {
	hash := hashToken(token)
	_, err := db.conn.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hash)
	return err
}

// DeleteUserSessions removes all sessions for a user (logout everywhere).
func (db *DB) DeleteUserSessions(userID int64) error {
	_, err := db.conn.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// CleanExpiredSessions removes all expired sessions.
func (db *DB) CleanExpiredSessions() error {
	_, err := db.conn.Exec(`DELETE FROM sessions WHERE expires_at < datetime('now')`)
	return err
}

// --- Submission Repository ---

// CreateSubmission records a new submission.
func (db *DB) CreateSubmission(userID int64, subType, name string, score int, cid, txHash string) (*Submission, error) {
	result, err := db.conn.Exec(`
		INSERT INTO submissions (user_id, type, name, score, cid, tx_hash) VALUES (?, ?, ?, ?, ?, ?)
	`, userID, subType, name, score, cid, txHash)
	if err != nil {
		return nil, fmt.Errorf("database: create submission failed: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetSubmissionByID(id)
}

// GetSubmissionByID retrieves a submission by ID.
func (db *DB) GetSubmissionByID(id int64) (*Submission, error) {
	var s Submission
	err := db.conn.QueryRow(`
		SELECT id, user_id, type, name, score, cid, tx_hash, created_at
		FROM submissions WHERE id = ?
	`, id).Scan(&s.ID, &s.UserID, &s.Type, &s.Name, &s.Score, &s.CID, &s.TxHash, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database: get submission failed: %w", err)
	}
	return &s, nil
}

// ListSubmissions returns submissions with pagination, optionally filtered by user.
func (db *DB) ListSubmissions(userID int64, limit, offset int) ([]*Submission, error) {
	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = db.conn.Query(`
			SELECT id, user_id, type, name, score, cid, tx_hash, created_at
			FROM submissions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
		`, userID, limit, offset)
	} else {
		rows, err = db.conn.Query(`
			SELECT id, user_id, type, name, score, cid, tx_hash, created_at
			FROM submissions ORDER BY created_at DESC LIMIT ? OFFSET ?
		`, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("database: list submissions failed: %w", err)
	}
	defer rows.Close()

	var subs []*Submission
	for rows.Next() {
		var s Submission
		if err := rows.Scan(&s.ID, &s.UserID, &s.Type, &s.Name, &s.Score, &s.CID, &s.TxHash, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

// CountSubmissions returns the total number of submissions, optionally filtered by user.
func (db *DB) CountSubmissions(userID int64) (int64, error) {
	var count int64
	var err error
	if userID > 0 {
		err = db.conn.QueryRow(`SELECT COUNT(*) FROM submissions WHERE user_id = ?`, userID).Scan(&count)
	} else {
		err = db.conn.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&count)
	}
	return count, err
}

// --- Helpers ---

// generateSessionToken creates a random token and HMAC-signs it.
// Format: <random_hex>.<hmac_hex>
// If SESSION_SECRET is not set, returns unsigned token (dev mode).
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	payload := hex.EncodeToString(b)
	if sessionSecret == "" {
		return payload
	}
	sig := signPayload(payload)
	return payload + "." + sig
}

// hashToken produces the SHA-256 hash used for database storage.
// Only the payload portion is hashed (signature is for tamper detection).
func hashToken(token string) string {
	payload := tokenPayload(token)
	h := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(h[:])
}

// ValidateToken checks the HMAC signature of a session token.
// Returns false if the token is tampered. Returns true if SESSION_SECRET is unset (dev mode).
func ValidateToken(token string) bool {
	if sessionSecret == "" {
		return true
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	expected := signPayload(parts[0])
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

func signPayload(payload string) string {
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func tokenPayload(token string) string {
	if idx := strings.IndexByte(token, '.'); idx != -1 {
		return token[:idx]
	}
	return token
}
