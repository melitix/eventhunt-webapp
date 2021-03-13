package main

import (
	"net/http"
)

/*
 * middlewareLOO is a middleware to protect routes that should only be accessed
 * by Logged Out Only users.
 *
 * Any users found logged in get a blind redirect to the homepage. Being
 * "logged in" is determined by the cookie session.
 */
func (a *app) middlewareLOO(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the user ID from the session. This allows us to determine if the
		// user is currently logged in.
		session, _ := store.Get(r, "login")
		userID, ok := session.Values["uid"].(uint64)

		// If user is logged in, redirect them to the homepage.
		if ok && userID != 0 {
			http.Redirect(w, r, "/", 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}
