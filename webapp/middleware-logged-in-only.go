package main

import (
	"log/slog"
	"net/http"

	"github.com/eventhunt-org/webapp/framework"
)

/*
 * middlewareLIO is a middleware that protects routes that should only be
 * accessed by authenticated users.
 */
func (a *app) middlewareLIO(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Context().Value("user") == nil {

			slog.Error("middleware: non-logged in user tried accessing an LIO route.")

			session, _ := store.Get(r, "login")
			session.AddFlash(framework.Flash{
				framework.FlashFail,
				"You must be logged in to view this page.",
			})
			err := session.Save(r, w)
			if err != nil {
				respondWithError(w, 502, err.Error())
				return
			}

			http.Redirect(w, r, "/login", 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}
