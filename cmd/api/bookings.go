package main

import (
	"errors"
	"fmt"
	"net/http"

	"thecodephilic-guy/eventbox/internal/data"
)

func (app *application) createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EventID int64 `json:"event_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r) // We know they are a customer due to requireRole middleware

	booking := &data.Booking{
		EventID:    input.EventID,
		CustomerID: user.ID,
	}

	// The Insert method handles the DB transaction: checking available tickets and decrementing
	err = app.models.Bookings.Insert(booking)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrSoldOut):
			app.errorResponse(w, r, http.StatusConflict, "sorry, this event is completely sold out")
		case errors.Is(err, data.ErrAlreadyBooked):
			app.errorResponse(w, r, http.StatusConflict, "you have already booked a ticket for this event")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// ==========================================
	// BACKGROUND TASK 1: Booking Confirmation
	// ==========================================
	app.background(func() {
		// Simulating sending a confirmation email to the customer
		app.logger.PrintInfo(fmt.Sprintf("BACKGROUND TASK: Sending booking confirmation email to Customer ID %d for Event ID %d", user.ID, booking.EventID), nil)
	})
	// ==========================================

	err = app.writeJSON(w, http.StatusCreated, envelop{"booking": booking}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBookingsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	bookings, err := app.models.Bookings.GetForCustomer(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"bookings": bookings}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
