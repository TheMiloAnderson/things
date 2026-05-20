package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"things/internal/db"
)

type TokenPurpose string

const (
	PurposeVerify TokenPurpose = "verify"
	PurposeReset  TokenPurpose = "reset"
)

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("expired token")
	ErrTokenUsed    = errors.New("token already used")
)

type AuthToken struct {
	db.Connection
	ID        int
	UserID    int
	TokenHash string
	Purpose   TokenPurpose
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

func hashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func generateTokenPlaintext() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func IssueToken(conn db.Connection, userID int, purpose TokenPurpose, ttl time.Duration) (plaintext string, err error) {
	plaintext, err = generateTokenPlaintext()
	if err != nil {
		return "", err
	}
	now := time.Now()
	t := AuthToken{
		Connection: conn,
		UserID:     userID,
		TokenHash:  hashToken(plaintext),
		Purpose:    purpose,
		ExpiresAt:  now.Add(ttl),
		CreatedAt:  now,
	}
	_, err = t.Exec(
		`INSERT INTO auth_tokens (user_id, token_hash, purpose, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		t.UserID, t.TokenHash, string(t.Purpose), t.ExpiresAt, t.CreatedAt,
	)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

func ConsumeToken(conn db.Connection, plaintext string, purpose TokenPurpose) (userID int, err error) {
	if plaintext == "" {
		return 0, ErrTokenInvalid
	}
	hash := hashToken(plaintext)

	if err := conn.BeginTx(); err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = conn.Rollback()
		}
	}()

	row := conn.QueryRow(
		`SELECT id, user_id, purpose, expires_at, used_at
		FROM auth_tokens WHERE token_hash = ? FOR UPDATE`,
		hash,
	)
	var t AuthToken
	var purposeStr string
	if err := row.Scan(&t.ID, &t.UserID, &purposeStr, &t.ExpiresAt, &t.UsedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTokenInvalid
		}
		return 0, err
	}
	t.Purpose = TokenPurpose(purposeStr)

	if t.Purpose != purpose {
		return 0, ErrTokenInvalid
	}
	if t.UsedAt.Valid {
		return 0, ErrTokenUsed
	}
	if time.Now().After(t.ExpiresAt) {
		return 0, ErrTokenExpired
	}

	now := time.Now()
	res, err := conn.Exec(`UPDATE auth_tokens SET used_at = ? WHERE id = ? AND used_at IS NULL`, now, t.ID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, ErrTokenUsed
	}

	if err := conn.Commit(); err != nil {
		return 0, err
	}
	return t.UserID, nil
}

func InvalidateUserTokens(conn db.Connection, userID int, purpose TokenPurpose) error {
	_, err := conn.Exec(
		`UPDATE auth_tokens SET used_at = ? WHERE user_id = ? AND purpose = ? AND used_at IS NULL`,
		time.Now(), userID, string(purpose),
	)
	return err
}

func CountRecentTokens(conn db.Connection, userID int, purpose TokenPurpose, since time.Time) (int, error) {
	row := conn.QueryRow(
		`SELECT COUNT(*) FROM auth_tokens
		WHERE user_id = ? AND purpose = ? AND created_at >= ?`,
		userID, string(purpose), since,
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func CountRecentTokensByEmail(conn db.Connection, email string, purpose TokenPurpose, since time.Time) (int, error) {
	row := conn.QueryRow(
		`SELECT COUNT(*) FROM auth_tokens t
		INNER JOIN users u ON u.id = t.user_id
		WHERE u.email = ? AND t.purpose = ? AND t.created_at >= ?`,
		email, string(purpose), since,
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, fmt.Errorf("count tokens by email: %w", err)
	}
	return n, nil
}
