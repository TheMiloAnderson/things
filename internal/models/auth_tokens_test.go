package models

import (
	"testing"
	"time"
)

func beginTokenTransaction(t *testing.T) (User, func()) {
	t.Helper()
	u := beginUserTransaction(t)
	return u, func() { _ = u.Rollback() }
}

func TestIssueAndConsumeToken(t *testing.T) {
	u, cleanup := beginTokenTransaction(t)
	defer cleanup()

	if err := u.GetById(1); err != nil {
		t.Fatalf("get user: %v", err)
	}

	plain, err := IssueToken(u.Connection, u.ID, PurposeVerify, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if plain == "" {
		t.Fatal("empty token")
	}

	gotID, err := ConsumeToken(u.Connection, plain, PurposeVerify)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if gotID != u.ID {
		t.Fatalf("user id %d want %d", gotID, u.ID)
	}

	_, err = ConsumeToken(u.Connection, plain, PurposeVerify)
	if err != ErrTokenUsed {
		t.Fatalf("replay: got %v want ErrTokenUsed", err)
	}
}

func TestConsumeTokenWrongPurpose(t *testing.T) {
	u, cleanup := beginTokenTransaction(t)
	defer cleanup()
	if err := u.GetById(1); err != nil {
		t.Fatalf("get user: %v", err)
	}

	plain, err := IssueToken(u.Connection, u.ID, PurposeVerify, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	_, err = ConsumeToken(u.Connection, plain, PurposeReset)
	if err != ErrTokenInvalid {
		t.Fatalf("got %v want ErrTokenInvalid", err)
	}
}

func TestConsumeTokenExpired(t *testing.T) {
	u, cleanup := beginTokenTransaction(t)
	defer cleanup()
	if err := u.GetById(1); err != nil {
		t.Fatalf("get user: %v", err)
	}

	plain, err := IssueToken(u.Connection, u.ID, PurposeVerify, -time.Second)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	_, err = ConsumeToken(u.Connection, plain, PurposeVerify)
	if err != ErrTokenExpired {
		t.Fatalf("got %v want ErrTokenExpired", err)
	}
}
