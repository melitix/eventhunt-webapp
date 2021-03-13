package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

/*
 * Routes are divided into many groups depending on the resource thy're working
 * with, the type of user, etc. Starting from the top however, there are three
 * main groups. Logged-out only (LOO), Logged-in only (LIO), and All Users
 * Welcomed (AUW).
 *
 * LOO - pages that should only be for unauthenticated users. These are pages
 * such as the signup/login pages, forgot password, etc.
 *
 * LIO - pages that should only be used with authenticated users. These are
 * typically pages that do CRUD operations such as creating groups or events,
 * deleting things, etc.
 *
 * AUW - these are pages that can be served to both logged in and not logged in
 * users. Unlike many typical webapps, this section is quite big with EventHunt.
 */
func (a *app) initializeRoutes() {

	// LOO pages
	a.Router.Group(func(r chi.Router) {

		r.Use(a.middlewareLOO)

		r.Get("/signup", a.authSignup)
		r.Post("/signup", a.authSignupPost)
		r.Get("/login", a.authLogin)
		r.Post("/login", a.authLoginPost)
		r.Get("/forgot-password", a.authForgotPasswordGet)
		r.Post("/forgot-password", a.authForgotPasswordPost)
		r.Get("/reset-password", a.resetPasswordGet)
		r.Post("/reset-password", a.resetPasswordPost)
	})

	// LIO & AUW pages
	a.Router.Group(func(r chi.Router) {

		r.Use(a.middlewareUser)

		r.Get("/", a.homepage)

		// Events
		r.Route("/events", func(r chi.Router) {
			r.Get("/", a.eventsIndex)
			r.With(a.middlewareLIO).Get("/{:new|schedule}", a.eventsNewAlias)
			r.With(a.middlewareLIO).Post("/{:new|schedule}", a.eventsNewPost)
			r.Route("/{event-id:[0-9]+}", func(r chi.Router) {
				r.Use(a.middlewareEvent)
				r.Get("/", a.eventsSingle)
				r.With(a.middlewareLIO).Get("/new-venue", a.venueNew)
				r.With(a.middlewareLIO).Post("/new-venue", a.venueNewPost)
				r.With(a.middlewareLIO).Get("/rsvp/{status:yes|maybe|no}", a.rsvpsInput)
			})
		})

		// Groups
		r.Route("/groups", func(r chi.Router) {
			r.Get("/", a.groupsIndex)
			r.Get("/{mode:all|my}", a.groupsIndex)
			r.With(a.middlewareLIO).Get("/new", a.groupsNew)
			r.With(a.middlewareLIO).Post("/new", a.groupsNewPost)
			r.Route("/{group-id:[0-9]+}", func(r chi.Router) {
				r.Use(a.middlewareGroup)
				r.Get("/", a.groupsSingle)
				r.Get("/{:new|schedule}", a.eventsNew)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(a.middlewareLIO)
			// For random pages
			r.Get("/invite", a.otherInviteGet)
			r.Post("/invite", a.otherInvitePost)
			r.Get("/verify-email", a.utilVerifyEmail)
		})
	})

	// This one is outside of LIO because sometimes for debugging purposes, we
	// have a cookie but no user. This allows us to reset the session basically.
	a.Router.Get("/logout", a.authLogout)
	a.Router.NotFound(http.HandlerFunc(a.util404Get))
}
