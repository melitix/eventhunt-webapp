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
 * middlewareGroup is a middleware that covers routes based on a single Group,
 * which uses the group ID.
 */
func (a *app) middlewareGroup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		gIDStr := chi.URLParam(r, "group-id")
		gID, err := strconv.ParseUint(gIDStr, 10, 64)
		if err != nil {
			slog.Error("middleware: Failed to parse ID.", "group-id", gIDStr, "err", err)
			respondWithError(w, 400, err.Error())
			return
		}

		g, err := db.GetGroupByID(a.DB, gID)
		if err != nil {
			slog.Error("middleware: Failed to load group from DB.", "id", gID, "err", err)
			respondWithError(w, 500, err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "group", g)

		// If the group is public, then we can just move on. Otherwise we need to
		// check permissions.
		if !g.IsPrivate {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if r.Context().Value("user") == nil {
			respondWithError(w, 403, "User does not have permission.")
			return
		}

		u := r.Context().Value("user").(*db.User)

		// check for permission
		if g.UserID != u.ID {
			respondWithError(w, 403, "User does not have permission.")
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
