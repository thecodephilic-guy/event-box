package validator

import (
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	v := New()
	if v.Errors == nil {
		t.Fatal("expected Errors map to be initialized, got nil")
	}
	if len(v.Errors) != 0 {
		t.Fatalf("expected empty Errors map, got %d entries", len(v.Errors))
	}
}

func TestValid(t *testing.T) {
	v := New()
	if !v.Valid() {
		t.Fatal("expected Valid() to return true for new validator")
	}

	v.AddError("field", "error message")
	if v.Valid() {
		t.Fatal("expected Valid() to return false after adding an error")
	}
}

func TestAddError(t *testing.T) {
	v := New()

	v.AddError("email", "must be provided")
	if v.Errors["email"] != "must be provided" {
		t.Fatalf("expected error 'must be provided', got %q", v.Errors["email"])
	}

	// Adding a second error for the same key should not overwrite the first.
	v.AddError("email", "must be valid")
	if v.Errors["email"] != "must be provided" {
		t.Fatalf("expected first error to be retained, got %q", v.Errors["email"])
	}
}

func TestCheck(t *testing.T) {
	v := New()

	// Passing check should not add an error.
	v.Check(true, "name", "must be provided")
	if !v.Valid() {
		t.Fatal("expected no errors after a passing check")
	}

	// Failing check should add an error.
	v.Check(false, "name", "must be provided")
	if v.Valid() {
		t.Fatal("expected an error after a failing check")
	}
	if v.Errors["name"] != "must be provided" {
		t.Fatalf("expected error 'must be provided', got %q", v.Errors["name"])
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		list     []string
		expected bool
	}{
		{"value present", "admin", []string{"admin", "user"}, true},
		{"value absent", "guest", []string{"admin", "user"}, false},
		{"empty list", "admin", []string{}, false},
		{"empty value in list", "", []string{"", "user"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := In(tt.value, tt.list...)
			if result != tt.expected {
				t.Errorf("In(%q, %v) = %v, want %v", tt.value, tt.list, result, tt.expected)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	rx := regexp.MustCompile(`^[a-z]+$`)

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"match", "hello", true},
		{"no match", "Hello", false},
		{"empty", "", false},
		{"numbers", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.value, rx); got != tt.expected {
				t.Errorf("Matches(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected bool
	}{
		{"all unique", []string{"a", "b", "c"}, true},
		{"duplicates", []string{"a", "b", "a"}, false},
		{"empty", []string{}, true},
		{"single", []string{"a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.values); got != tt.expected {
				t.Errorf("Unique(%v) = %v, want %v", tt.values, got, tt.expected)
			}
		})
	}
}

func TestEmailRX(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name+tag@example.co.uk", true},
		{"@example.com", false},
		{"plaintext", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := Matches(tt.email, EmailRX); got != tt.valid {
				t.Errorf("EmailRX.Match(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}
