package data

import (
	"testing"

	"thecodephilic-guy/eventbox/internal/validator"
)

func TestPasswordSetAndMatches(t *testing.T) {
	p := &password{}
	err := p.Set("validpassword123")
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	if p.plaintext == nil {
		t.Fatal("expected plaintext to be set")
	}
	if *p.plaintext != "validpassword123" {
		t.Errorf("expected plaintext 'validpassword123', got %q", *p.plaintext)
	}
	if p.hash == nil {
		t.Fatal("expected hash to be set")
	}

	// Correct password should match.
	match, err := p.Matches("validpassword123")
	if err != nil {
		t.Fatalf("Matches() returned error: %v", err)
	}
	if !match {
		t.Error("expected Matches() to return true for correct password")
	}

	// Wrong password should not match.
	match, err = p.Matches("wrongpassword")
	if err != nil {
		t.Fatalf("Matches() returned error: %v", err)
	}
	if match {
		t.Error("expected Matches() to return false for wrong password")
	}
}

func TestIsAnonymous(t *testing.T) {
	if !AnonymousUser.IsAnonymous() {
		t.Error("expected AnonymousUser.IsAnonymous() to return true")
	}

	user := &User{ID: 1, Name: "Test"}
	if user.IsAnonymous() {
		t.Error("expected non-anonymous user.IsAnonymous() to return false")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "test@example.com", true},
		{"empty email", "", false},
		{"invalid email", "not-an-email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateEmail(v, tt.email)
			if v.Valid() != tt.valid {
				t.Errorf("ValidateEmail(%q) valid = %v, want %v (errors: %v)", tt.email, v.Valid(), tt.valid, v.Errors)
			}
		})
	}
}

func TestValidatePasswordPlaintext(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid", "securepassword123", true},
		{"empty", "", false},
		{"too short", "short", false},
		{"exactly 8", "12345678", true},
		{"73 chars", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePasswordPlaintext(v, tt.password)
			if v.Valid() != tt.valid {
				t.Errorf("ValidatePasswordPlaintext(%q) valid = %v, want %v (errors: %v)", tt.password, v.Valid(), tt.valid, v.Errors)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	validUser := func() *User {
		u := &User{
			Name:  "Sohail",
			Email: "sohail@example.com",
			Role:  "customer",
		}
		u.Password.Set("securepassword123")
		return u
	}

	t.Run("valid user", func(t *testing.T) {
		v := validator.New()
		ValidateUser(v, validUser())
		if !v.Valid() {
			t.Errorf("expected valid, got errors: %v", v.Errors)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		u := validUser()
		u.Name = ""
		v := validator.New()
		ValidateUser(v, u)
		if v.Valid() {
			t.Error("expected validation error for empty name")
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		u := validUser()
		u.Role = "admin"
		v := validator.New()
		ValidateUser(v, u)
		if v.Valid() {
			t.Error("expected validation error for invalid role")
		}
	})

	t.Run("empty role", func(t *testing.T) {
		u := validUser()
		u.Role = ""
		v := validator.New()
		ValidateUser(v, u)
		if v.Valid() {
			t.Error("expected validation error for empty role")
		}
	})

	t.Run("organizer role", func(t *testing.T) {
		u := validUser()
		u.Role = "organizer"
		v := validator.New()
		ValidateUser(v, u)
		if !v.Valid() {
			t.Errorf("expected valid organizer, got errors: %v", v.Errors)
		}
	})
}
