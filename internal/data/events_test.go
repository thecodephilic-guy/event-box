package data

import (
	"testing"
	"time"

	"thecodephilic-guy/eventbox/internal/validator"
)

func TestValidateEvent(t *testing.T) {
	validEvent := func() *Event {
		return &Event{
			Title:            "Tech Conference 2027",
			Description:      "A great tech conference",
			Date:             time.Now().Add(24 * time.Hour),
			TotalTickets:     100,
			AvailableTickets: 100,
		}
	}

	t.Run("valid event", func(t *testing.T) {
		v := validator.New()
		ValidateEvent(v, validEvent())
		if !v.Valid() {
			t.Errorf("expected valid, got errors: %v", v.Errors)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		e := validEvent()
		e.Title = ""
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["title"]; !ok {
			t.Error("expected validation error for empty title")
		}
	})

	t.Run("title too long", func(t *testing.T) {
		e := validEvent()
		e.Title = string(make([]byte, 256))
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["title"]; !ok {
			t.Error("expected validation error for title too long")
		}
	})

	t.Run("empty description", func(t *testing.T) {
		e := validEvent()
		e.Description = ""
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["description"]; !ok {
			t.Error("expected validation error for empty description")
		}
	})

	t.Run("zero date", func(t *testing.T) {
		e := validEvent()
		e.Date = time.Time{}
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["date"]; !ok {
			t.Error("expected validation error for zero date")
		}
	})

	t.Run("zero total_tickets", func(t *testing.T) {
		e := validEvent()
		e.TotalTickets = 0
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["total_tickets"]; !ok {
			t.Error("expected validation error for zero total_tickets")
		}
	})

	t.Run("negative available_tickets", func(t *testing.T) {
		e := validEvent()
		e.AvailableTickets = -1
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["available_tickets"]; !ok {
			t.Error("expected validation error for negative available_tickets")
		}
	})

	t.Run("available exceeds total", func(t *testing.T) {
		e := validEvent()
		e.AvailableTickets = 200
		v := validator.New()
		ValidateEvent(v, e)
		if _, ok := v.Errors["available_tickets"]; !ok {
			t.Error("expected validation error for available > total")
		}
	})

	// ValidateEvent (for updates) should NOT reject a past date.
	t.Run("past date allowed in ValidateEvent", func(t *testing.T) {
		e := validEvent()
		e.Date = time.Now().Add(-24 * time.Hour) // past date
		v := validator.New()
		ValidateEvent(v, e)
		if !v.Valid() {
			t.Errorf("ValidateEvent should allow past dates (for updates), got errors: %v", v.Errors)
		}
	})
}

func TestValidateNewEvent(t *testing.T) {
	t.Run("past date rejected", func(t *testing.T) {
		e := &Event{
			Title:            "Past Event",
			Description:      "Should fail",
			Date:             time.Now().Add(-24 * time.Hour),
			TotalTickets:     100,
			AvailableTickets: 100,
		}
		v := validator.New()
		ValidateNewEvent(v, e)
		if _, ok := v.Errors["date"]; !ok {
			t.Error("ValidateNewEvent should reject past dates")
		}
	})

	t.Run("future date accepted", func(t *testing.T) {
		e := &Event{
			Title:            "Future Event",
			Description:      "Should pass",
			Date:             time.Now().Add(24 * time.Hour),
			TotalTickets:     100,
			AvailableTickets: 100,
		}
		v := validator.New()
		ValidateNewEvent(v, e)
		if !v.Valid() {
			t.Errorf("ValidateNewEvent should accept future dates, got errors: %v", v.Errors)
		}
	})
}
