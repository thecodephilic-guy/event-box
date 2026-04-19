package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"thecodephilic-guy/eventbox/internal/data"
	"thecodephilic-guy/eventbox/internal/validator"
)

func (app *application) createEventHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string    `json:"title"`
		Description  string    `json:"description"`
		Date         time.Time `json:"date"` // Expects RFC3339 string like "2026-12-01T10:00:00Z"
		TotalTickets int       `json:"total_tickets"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r) // We know they are an organizer due to requireRole middleware

	event := &data.Event{
		OrganizerID:      user.ID,
		Title:            input.Title,
		Description:      input.Description,
		Date:             input.Date,
		TotalTickets:     input.TotalTickets,
		AvailableTickets: input.TotalTickets, // Initially, all tickets are available
	}

	v := validator.New()
	if data.ValidateNewEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Events.Insert(event)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/events/%d", event.ID))

	err = app.writeJSON(w, http.StatusCreated, envelop{"event": event}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParams(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"event": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParams(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Security Check: Only the organizer who created the event can update it
	user := app.contextGetUser(r)
	if event.OrganizerID != user.ID {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Date        *time.Time `json:"date"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Description != nil {
		event.Description = *input.Description
	}
	if input.Date != nil {
		event.Date = *input.Date
	}

	v := validator.New()
	if data.ValidateEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Events.Update(event)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// ==========================================
	// BACKGROUND TASK 2: Event Update Notification
	// ==========================================
	app.background(func() {
		// Simulating sending an email to all customers who booked this event
		app.logger.PrintInfo(fmt.Sprintf("BACKGROUND TASK: Sending update notification to all customers for Event ID: %d ('%s')", event.ID, event.Title), nil)
	})
	// ==========================================

	err = app.writeJSON(w, http.StatusOK, envelop{"event": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "date", "-id", "-title", "-date"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	events, metadata, err := app.models.Events.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"events": events, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
