package main

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func loginUserHelper(w http.ResponseWriter, r *http.Request) {

	if userCreated {
		return
	}

	// create a test user
	user, err := createUser(a.DB, "MrJackson", "Bumpty", "jackson@halsey.edu", "Paul", "Jackson")
	if err != nil {
		log.Fatal(err)
	}

	session, _ := store.Get(r, "login")

	session.Values["authenticated"] = true
	session.Values["uid"] = user.ID
	session.Save(r, w)

	userCreated = true
}

/*
 * Manually injects variables into the context that would normally be handled
 * by middleware. In handler unit testing, middleware doesn't run hence why this
 * is needed.
 */
func middlewareTestHelper(r *http.Request, mode string) *http.Request {

	ctx := r.Context()
	ctx = context.WithValue(ctx, "brandMode", mode)

	return r.WithContext(ctx)
}
