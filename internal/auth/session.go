package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// SessionStore manages refresh token sessions in SQLite.
type SessionStore struct {
	db *sql.DB
}

// NewSessionStore creates a new SessionStore.
func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

// HashToken hashes a token with SHA-256.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateSession stores a new refresh token session.
func (s *SessionStore) CreateSession(ctx context.Context, userID int64, token, ip, userAgent string, expiresAt time.Time, csrfTokenHash string) error {
	hash := HashToken(token)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at, csrf_token_hash) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, hash, ip, userAgent, expiresAt, csrfTokenHash,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// SetCSRFTokenHash updates the CSRF token hash for a session.
func (s *SessionStore) SetCSRFTokenHash(ctx context.Context, token string, csrfHash string) error {
	hash := HashToken(token)
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET csrf_token_hash = ? WHERE token_hash = ?`,
		csrfHash, hash,
	)
	if err != nil {
		return fmt.Errorf("set CSRF token hash: %w", err)
	}
	return nil
}

// GetCSRFTokenHash retrieves the CSRF token hash for the session.
func (s *SessionStore) GetCSRFTokenHash(ctx context.Context, token string) (string, error) {
	hash := HashToken(token)
	var csrfHash string
	row := s.db.QueryRowContext(ctx, `SELECT csrf_token_hash FROM sessions WHERE token_hash = ?`, hash)
	if err := row.Scan(&csrfHash); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get CSRF token hash: %w", err)
	}
	return csrfHash, nil
}

// ValidateSession checks if a refresh token is valid and not revoked/expired.
func (s *SessionStore) ValidateSession(ctx context.Context, token string) (userID int64, err error) {
	hash := HashToken(token)
	var id int64
	var expiresAt time.Time
	var revoked bool
	row := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at, revoked FROM sessions WHERE token_hash = ?`, hash)
	if err := row.Scan(&id, &expiresAt, &revoked); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("session not found")
		}
		return 0, fmt.Errorf("validate session: %w", err)
	}
	if revoked {
		return 0, fmt.Errorf("session revoked")
	}
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("session expired")
	}
	return id, nil
}

// RevokeSession marks a session as revoked.
func (s *SessionStore) RevokeSession(ctx context.Context, sessionID int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET revoked = TRUE WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user except the current one.
func (s *SessionStore) RevokeAllUserSessions(ctx context.Context, userID int64, exceptToken string) error {
	exceptHash := HashToken(exceptToken)
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET revoked = TRUE WHERE user_id = ? AND token_hash != ?`,
		userID, exceptHash,
	)
	if err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}
	return nil
}

// ListUserSessions returns all active sessions for a user.
func (s *SessionStore) ListUserSessions(ctx context.Context, userID int64) ([]SessionInfo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ip_address, user_agent, created_at, expires_at FROM sessions WHERE user_id = ? AND revoked = FALSE AND expires_at > ?`,
		userID, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var si SessionInfo
		var csrfHash sql.NullString
		if err := rows.Scan(&si.ID, &si.IPAddress, &si.UserAgent, &si.CreatedAt, &si.ExpiresAt, &csrfHash); err != nil {
			if err := rows.Scan(&si.ID, &si.IPAddress, &si.UserAgent, &si.CreatedAt, &si.ExpiresAt); err != nil {
				return nil, fmt.Errorf("scan session: %w", err)
			}
		} else {
			si.CSRFTokenHash = csrfHash.String
		}
		sessions = append(sessions, si)
	}
	return sessions, rows.Err()
}

// SessionInfo represents session metadata.
type SessionInfo struct {
	ID            int64     `json:"id"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	CSRFTokenHash string    `json:"-"`
}

// CleanupExpiredSessions removes expired sessions.
func (s *SessionStore) CleanupExpiredSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}
	return nil
}
