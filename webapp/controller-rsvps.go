package main

import (
	"log/slog"
	"net/http"

	"github.com/eventhunt-org/webapp/webapp/db"
	"github.com/jackc/pgx/v5"

	"github.com/go-chi/chi/v5"
)

/*
 * Save an RSVP status.
 */
func (a *app) rsvpsInput(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a User
	u := r.Context().Value("user").(*db.User)
	// middlewareEvent ensures we have an Event
	e := r.Context().Value("event").(*db.Event)

	rsvpIntent := db.RSVPStatus(chi.URLParam(r, "status"))

	// check if we have an existing RSVP first
	rsvp, err := db.GetRSVP(a.DB, e.ID, u.ID)
	if err != nil && err != pgx.ErrNoRows {
		slog.Error("Failed to check RSVP status.", "err", err)
		session.Values["message"] = "Failed to RSVP."
		session.Save(r, w)

		http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
		return
	}

	// we have an RSVP so let's just update it
	if rsvp != nil {
		rsvp.Intent = rsvpIntent
		rsvp.DB = a.DB
		rsvp.Save()
	} else { // let's create a new one
		_, err := db.NewRSVP(e.ID, u, rsvpIntent, db.RSVPAttendee)
		if err != nil {
			slog.Error("Failed to RSVP.", "err", err)
			session.Values["message"] = "Failed to RSVP."
			session.Save(r, w)
		}
	}

	http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
	return
}
