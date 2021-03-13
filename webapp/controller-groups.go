package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/go-chi/chi/v5"
)

/*
 * Handles the groups index page.
 *
 * Path: /groups
 */
func (a *app) groupsIndex(w http.ResponseWriter, r *http.Request) {

	// middlewareUser might provide a User
	u, ok := r.Context().Value("user").(*db.User)

	mode := chi.URLParam(r, "mode")
	var groups []*db.Group
	var err error

	if (mode == "" || mode == "my") && ok {

		mode = "my"
		groups, err = db.GetGroupsByUser(u)
		if err != nil {
			slog.Error("Failed to get the list of user's groups.", "err", err)
		}
	} else {

		groups, err = db.GetGroupsByLimit(a.DB, 25)
		if err != nil {
			slog.Error("Failed to get a list of groups.", "err", err)
		}
	}

	renderPage(a, "groups/index.html", w, r, map[string]interface{}{
		"User":   u,
		"Groups": groups,
		"Mode":   mode,
	})
}

/*
 * View a single event.
 */
func (a *app) groupsSingle(w http.ResponseWriter, r *http.Request) {

	// middlewareUser might have provided us a User
	u, _ := r.Context().Value("user").(*db.User)
	// middlewareGroup ensures we have a Group
	g := r.Context().Value("group").(*db.Group)

	renderPage(a, "groups/single.tmpl", w, r, map[string]interface{}{
		"User":  u,
		"Group": g,
	})
}

func (this *app) GeventsEditGet(w http.ResponseWriter, r *http.Request) {
}

func (a *app) GeventsEditPost(w http.ResponseWriter, r *http.Request) {

}

/*
 * Handles the page to create a new group.
 *
 * Path: /groups/new
 */
func (a *app) groupsNew(w http.ResponseWriter, r *http.Request) {

	var messages []string

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a user
	u := r.Context().Value("user").(*db.User)

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

		http.Redirect(w, r, "/groups", http.StatusFound)
		return
	}

	renderPage(a, "groups/new.html", w, r, map[string]interface{}{
		"Messages": messages,
		"User":     u,
		"Cities":   cities,
	})
}

/*
 * Processes the page to create a new group.
 *
 * Path: /groups/new
 */
func (a *app) groupsNewPost(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")
	u, _ := initUser(a, r)

	r.ParseForm()
	defer r.Body.Close()

	name := r.Form.Get("group-name")
	city := r.Form.Get("city")
	visibility := r.Form.Get("visibility")

	cityID, err := strconv.ParseUint(city, 10, 64)
	if err != nil {

		slog.Error("City ID is not valid.", "id", city)
		session.Values["message"] = "City was invalid."
		session.Save(r, w)

		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	var isPrivate bool
	switch visibility {
	case "public":
		isPrivate = false
	case "private":
		isPrivate = true
	default:

		slog.Error("Visibility was invalid.", "visibility", visibility)

		session.Values["message"] = "Visibility was invalid."
		session.Save(r, w)

		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	_, err = db.NewGroup(u, name, cityID, isPrivate)
	if err != nil {

		slog.Error("Failed to create group.", "err", err)

		session.Values["message"] = "Failed to create group."
		session.Save(r, w)

		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/groups", http.StatusFound)
	return
}
