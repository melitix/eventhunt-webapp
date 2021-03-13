package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/go-chi/chi/v5"
)

/*
 * middlewareEvent is a middleware that covers routes based on a single Event,
 * which uses the event ID.
 */
func (a *app) middlewareEvent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		eIDStr := chi.URLParam(r, "event-id")
		eID, err := strconv.ParseUint(eIDStr, 10, 64)
		if err != nil {
			slog.Error("middleware: Failed to parse ID.", "event-id", eIDStr, "err", err)
			respondWithError(w, 400, err.Error())
			return
		}

		e, err := db.GetEventByID(a.DB, eID)
		if err != nil {
			slog.Error("middleware: Failed to load event from DB.", "id", eID, "err", err)
			respondWithError(w, 500, err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "event", e)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
