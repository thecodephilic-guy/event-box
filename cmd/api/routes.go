package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// User & Auth Routes
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Event Routes
	// Both customers and organizers can list/view events (must be authenticated)
	router.HandlerFunc(http.MethodGet, "/v1/events", app.requireAuthenticatedUser(app.listEventsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/events/:id", app.requireAuthenticatedUser(app.showEventHandler))

	// Only Organizers can manage events
	router.HandlerFunc(http.MethodPost, "/v1/events", app.requireRole("organizer", app.createEventHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/events/:id", app.requireRole("organizer", app.updateEventHandler))

	// Booking Routes
	// Only Customers can book tickets
	router.HandlerFunc(http.MethodPost, "/v1/bookings", app.requireRole("customer", app.createBookingHandler))

	// Optional: Customers can see their own bookings, Organizers can see bookings for their events
	router.HandlerFunc(http.MethodGet, "/v1/bookings", app.requireAuthenticatedUser(app.listBookingsHandler))

	// Metrics
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Middleware chain
	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
