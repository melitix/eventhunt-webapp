package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/eventhunt-org/webapp/webapp/db"
)

/*
 * middlewareUser is a middleware that loads a User struct into the context if
 * there is a user session (cookie).
 */
func (a *app) middlewareUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the user ID from the session. This allows us to determine if the
		// user is currently logged in.
		session, _ := store.Get(r, "login")
		ctx := r.Context()

		userID, ok := session.Values["uid"].(uint64)

		if ok && userID != 0 {

			// We got an ID, let's try to load that user
			u, err := db.GetUserByID(a.DB, userID)
			if err != nil {

				slog.Error("middleware: Failed to load user.", "id", userID, "err", err)
				return
			} else {

				// store the user and project to the context
				ctx = context.WithValue(ctx, "user", u)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
