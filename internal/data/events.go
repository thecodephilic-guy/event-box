package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"thecodephilic-guy/eventbox/internal/validator"
)

type Event struct {
	ID               int64     `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	OrganizerID      int64     `json:"organizer_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Date             time.Time `json:"date"`
	TotalTickets     int       `json:"total_tickets"`
	AvailableTickets int       `json:"available_tickets"`
	Version          int       `json:"version"`
}

// ValidateEvent performs general field validation (used for updates).
func ValidateEvent(v *validator.Validator, event *Event) {
	v.Check(event.Title != "", "title", "must be provided")
	v.Check(len(event.Title) <= 255, "title", "must not be more than 255 bytes long")

	v.Check(event.Description != "", "description", "must be provided")

	v.Check(!event.Date.IsZero(), "date", "must be a valid date")

	v.Check(event.TotalTickets > 0, "total_tickets", "must be greater than zero")
	v.Check(event.AvailableTickets >= 0, "available_tickets", "must not be negative")
	v.Check(event.AvailableTickets <= event.TotalTickets, "available_tickets", "cannot exceed total tickets")
}

// ValidateNewEvent performs full validation including the future-date constraint (used for creation).
func ValidateNewEvent(v *validator.Validator, event *Event) {
	ValidateEvent(v, event)
	v.Check(event.Date.After(time.Now()), "date", "must be in the future")
}

type EventModel struct {
	DB *sql.DB
}

func (m EventModel) Insert(event *Event) error {
	query := `
		INSERT INTO events (organizer_id, title, description, date, total_tickets, available_tickets)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, version`

	args := []any{event.OrganizerID, event.Title, event.Description, event.Date, event.TotalTickets, event.AvailableTickets}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&event.ID, &event.CreatedAt, &event.Version)
}

func (m EventModel) Get(id int64) (*Event, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, organizer_id, title, description, date, total_tickets, available_tickets, version
		FROM events
		WHERE id = $1`

	var event Event
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.CreatedAt, &event.OrganizerID, &event.Title,
		&event.Description, &event.Date, &event.TotalTickets, &event.AvailableTickets, &event.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &event, nil
}

func (m EventModel) Update(event *Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, date = $3, total_tickets = $4, available_tickets = $5, version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version`

	args := []any{
		event.Title, event.Description, event.Date,
		event.TotalTickets, event.AvailableTickets, event.ID, event.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&event.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}
	return nil
}

func (m EventModel) GetAll(title string, filters Filters) ([]*Event, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, organizer_id, title, description, date, total_tickets, available_tickets, version
		FROM events
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	events := []*Event{}

	for rows.Next() {
		var event Event
		err := rows.Scan(
			&totalRecords, &event.ID, &event.CreatedAt, &event.OrganizerID,
			&event.Title, &event.Description, &event.Date,
			&event.TotalTickets, &event.AvailableTickets, &event.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return events, metadata, nil
}
