package data

import (
	"testing"

	"thecodephilic-guy/eventbox/internal/validator"
)

func TestGenerateToken(t *testing.T) {
	token, err := generateToken(1, 24*60*60*1e9, ScopeAuthentication) // 24 hours in nanoseconds
	if err != nil {
		t.Fatalf("generateToken() returned error: %v", err)
	}

	if token.Plaintext == "" {
		t.Error("expected non-empty plaintext token")
	}

	// base32 encoding of 16 bytes without padding = 26 characters
	if len(token.Plaintext) != 26 {
		t.Errorf("expected plaintext length 26, got %d", len(token.Plaintext))
	}

	if token.Hash == nil || len(token.Hash) == 0 {
		t.Error("expected non-empty hash")
	}

	// SHA-256 produces a 32-byte hash.
	if len(token.Hash) != 32 {
		t.Errorf("expected hash length 32, got %d", len(token.Hash))
	}

	if token.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", token.UserID)
	}

	if token.Scope != ScopeAuthentication {
		t.Errorf("expected scope %q, got %q", ScopeAuthentication, token.Scope)
	}

	if token.Expiry.IsZero() {
		t.Error("expected non-zero Expiry")
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	token1, _ := generateToken(1, 24*60*60*1e9, ScopeAuthentication)
	token2, _ := generateToken(1, 24*60*60*1e9, ScopeAuthentication)

	if token1.Plaintext == token2.Plaintext {
		t.Error("expected two generated tokens to be unique")
	}
}

func TestValidateTokenPlaintext(t *testing.T) {
	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{"valid 26 chars", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", true},
		{"empty", "", false},
		{"too short", "ABCDE", false},
		{"too long", "ABCDEFGHIJKLMNOPQRSTUVWXYZ1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateTokenPlaintext(v, tt.token)
			if v.Valid() != tt.valid {
				t.Errorf("ValidateTokenPlaintext(%q) valid = %v, want %v (errors: %v)", tt.token, v.Valid(), tt.valid, v.Errors)
			}
		})
	}
}
