package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/eventhunt-org/webapp/framework"
	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/go-chi/chi/v5"
)

func (a *app) eventsIndex(w http.ResponseWriter, r *http.Request) {

	// middlewareUser might give us a User
	u, _ := r.Context().Value("user").(*db.User)

	events, err := db.GetEvents(a.DB, 25)
	if err != nil {
		slog.Error("Failed to get a list of events.", "err", err)
	}

	renderPage(a, "events/index", w, r, map[string]interface{}{
		"User":   u,
		"Events": events,
	})
}

/*
 * View a single event.
 */
func (a *app) eventsSingle(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareUser might give us a User
	u, _ := r.Context().Value("user").(*db.User)

	eIDStr := chi.URLParam(r, "event-id")
	eID, err := strconv.ParseUint(eIDStr, 10, 64)
	if err != nil {

		slog.Error("Failed to create event.", "msg", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Failed to create event.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events", http.StatusFound)
		return
	}

	e, err := db.GetEventByID(a.DB, eID)
	if err != nil {

		slog.Error("Failed to retrieve event from DB.", "msg", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Failed to create event.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events", http.StatusFound)
		return
	}

	renderPage(a, "events/single", w, r, map[string]interface{}{
		"User":  u,
		"Event": e,
	})
}

func (this *app) eventsEditGet(w http.ResponseWriter, r *http.Request) {
}

func (a *app) eventsEditPost(w http.ResponseWriter, r *http.Request) {

}

/*
 * eventsNewAlias provides a shorthand for creating a new event in the user's
 * primary group. If the user doesn't have a group, redirects to the create new
 * group page.
 */
func (a *app) eventsNewAlias(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// LIO middleware ensures we have a user
	u := r.Context().Value("user").(*db.User)

	groups, err := db.GetGroupsByUser(u)
	if err != nil {

		slog.Error("Failed to get a list of groups.", "err", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"There was a problem loading your groups.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events", http.StatusFound)
		return

	}

	if len(groups) == 0 {

		session.AddFlash(framework.Flash{
			framework.FlashWarn,
			"Please create a group before trying to create an event.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/groups/"+groups[0].IDString()+"/schedule", http.StatusFound)
	return
}

/*
 * Page to create a new event.
 */
func (a *app) eventsNew(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a user
	u := r.Context().Value("user").(*db.User)
	// middlewareGroup ensures we have a Group
	g := r.Context().Value("group").(*db.Group)

	// Only the owner of the group can create an event
	if !g.HasCreate(u.ID) {

		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"You don't have permission to create events for this group.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/"+g.IDString(), http.StatusFound)
		return
	}

	renderPage(a, "events/new", w, r, map[string]interface{}{
		"User":  u,
		"Group": g,
	})
}

/*
 * Processing of the new event page.
 */
func (a *app) eventsNewPost(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a User
	u, _ := r.Context().Value("user").(*db.User)
	// middlewareGroup ensures we have a Group
	g := r.Context().Value("group").(*db.Group)

	// Only the owner of the group can create an event
	if !g.HasCreate(u.ID) {

		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"You don't have permission to create events for this group.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/"+g.IDString(), http.StatusFound)
		return
	}

	r.ParseForm()
	defer r.Body.Close()

	name := r.Form.Get("event-name")
	summary := r.Form.Get("event-summary")
	timezone := r.Form.Get("timezone")

	loc, err := time.LoadLocation(timezone)
	if err != nil {

		slog.Error("Timezone is not parsable.", "timezone", timezone)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Timezone was not valid.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events/new", http.StatusFound)
		return
	}

	startTime, err := time.ParseInLocation("2006-01-02T15:04", r.Form.Get("start-time"), loc)
	if err != nil {

		slog.Error("Start time is not parsable.", "start-time", r.Form.Get("start-time"))
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Start time was not valid.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events/new", http.StatusFound)
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02T15:04", r.Form.Get("end-time"), loc)
	if err != nil {

		slog.Error("End time is not parsable.", "end-time", r.Form.Get("end-time"))
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"End time was not valid.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events/new", http.StatusFound)
		return
	}

	e, err := db.NewEvent(u, g.ID, name, startTime, endTime, summary)
	if err != nil {

		slog.Error("Failed to create event.", "msg", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Failed to create event.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/events/new", http.StatusFound)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/events/"+e.IDString()+"/new-venue", http.StatusFound)
	return
}
