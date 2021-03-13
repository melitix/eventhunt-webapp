package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/eventhunt-org/webapp/webapp/db"
)

/*
 * Page to add a venue to an event.
 */
func (a *app) venueNew(w http.ResponseWriter, r *http.Request) {

	var messages []string

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a user
	u := r.Context().Value("user").(*db.User)
	// middlewareEvent ensures we have an Event
	e := r.Context().Value("event").(*db.Event)

	if session.Values["message"] != nil && session.Values["message"].(string) != "" {

		messages = append(messages, session.Values["message"].(string))
		session.Values["message"] = ""
		session.Save(r, w)
	}

	cities, err := db.GetCitiesByAll(a.DB)
	if err != nil {

		slog.Error("Failed to get all cities.", "err", err)
		session.Values["message"] = "Cities list failed to load."
		session.Save(r, w)

		http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
		return
	}

	renderPage(a, "venues/new.tmpl", w, r, map[string]interface{}{
		"Messages": messages,
		"User":     u,
		"Event":    e,
		"Cities":   cities,
	})
}

/*
 * Processing of the new venue event page.
 */
func (a *app) venueNewPost(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareEvent ensures we have an Event
	e := r.Context().Value("event").(*db.Event)

	r.ParseForm()
	defer r.Body.Close()

	name := r.Form.Get("venue-name")
	address := r.Form.Get("street-address")
	city := r.Form.Get("city")

	cityID, err := strconv.ParseUint(city, 10, 64)
	if err != nil {

		slog.Error("City ID is not valid.", "id", city)
		session.Values["message"] = "City was invalid."
		session.Save(r, w)

		http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
		return
	}

	v, err := db.NewVenue(a.DB, name, address, cityID)
	if err != nil {

		slog.Error("Failed to create venue.", "err", err)
		session.Values["message"] = "Failed to create venue."
		session.Save(r, w)

		http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
		return
	}

	e.Venue = v
	e.Save()

	http.Redirect(w, r, "/events/"+e.IDString(), http.StatusFound)
	return
}
