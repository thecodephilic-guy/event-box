package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrSoldOut       = errors.New("event is sold out")
	ErrAlreadyBooked = errors.New("you have already booked a ticket for this event")
)

type Booking struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	EventID    int64     `json:"event_id"`
	CustomerID int64     `json:"customer_id"`
}

type BookingModel struct {
	DB *sql.DB
}

// Insert creates a booking and decrements the available tickets atomically using a transaction.
func (m BookingModel) Insert(booking *Booking) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 1. Begin the Database Transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback. If the transaction is committed successfully, this is a no-op.
	defer tx.Rollback()

	// 2. Atomically decrement available tickets if they are greater than 0
	updateQuery := `
		UPDATE events 
		SET available_tickets = available_tickets - 1, version = version + 1 
		WHERE id = $1 AND available_tickets > 0`

	res, err := tx.ExecContext(ctx, updateQuery, booking.EventID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	// If no rows were updated, it means available_tickets was 0.
	if rowsAffected == 0 {
		return ErrSoldOut
	}

	// 3. Insert the Booking record
	insertQuery := `
		INSERT INTO bookings (event_id, customer_id) 
		VALUES ($1, $2) 
		RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, insertQuery, booking.EventID, booking.CustomerID).Scan(&booking.ID, &booking.CreatedAt)
	if err != nil {
		// Catch the unique constraint violation (Customer booking same event twice)
		if err.Error() == `pq: duplicate key value violates unique constraint "bookings_event_id_customer_id_key"` {
			return ErrAlreadyBooked
		}
		return err
	}

	// 4. Commit the transaction
	return tx.Commit()
}

// GetForCustomer retrieves all bookings made by a specific customer.
func (m BookingModel) GetForCustomer(customerID int64) ([]*Booking, error) {
	query := `
		SELECT id, created_at, event_id, customer_id
		FROM bookings
		WHERE customer_id = $1
		ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		var booking Booking
		err := rows.Scan(&booking.ID, &booking.CreatedAt, &booking.EventID, &booking.CustomerID)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &booking)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}
