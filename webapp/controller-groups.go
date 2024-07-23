package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/eventhunt-org/webapp/framework"
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

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a user
	u := r.Context().Value("user").(*db.User)

	cities, err := db.GetCitiesByAll(a.DB)
	if err != nil {

		slog.Error("Failed to get all cities.", "err", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Cities list failed to load.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups", http.StatusFound)
		return
	}

	renderPage(a, "groups/new.html", w, r, map[string]interface{}{
		"User":   u,
		"Cities": cities,
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
	summary := r.Form.Get("group-summary")

	cityID, err := strconv.ParseUint(city, 10, 64)
	if err != nil {

		slog.Error("City ID is not valid.", "id", city)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"City was invalid.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	isPrivate := false

	_, err = db.NewGroup(u, name, cityID, summary, isPrivate)
	if err != nil {

		slog.Error("Failed to create group.", "err", err)
		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Failed to create group.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/new", http.StatusFound)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/groups", http.StatusFound)
	return
}

/*
 * Handles joining a Group as a member.
 *
 * Path: /groups/join
 */
func (a *app) groupsJoin(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "login")

	// middlewareLIO ensures we have a User
	u := r.Context().Value("user").(*db.User)
	// middlewareGroup ensures we have a Group
	g := r.Context().Value("group").(*db.Group)

	// can't join a group we're already in
	if g.IsMember(u.ID) {

		session.AddFlash(framework.Flash{
			framework.FlashWarn,
			"Can't join a group you're already in.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/"+g.IDString(), http.StatusFound)
		return
	}

	// We're clear to join
	_, err := db.NewMembership(g.ID, u, db.MemberMember)
	if err != nil {

		slog.Error("Failed to create membership.", "err", err)

		session.AddFlash(framework.Flash{
			framework.FlashFail,
			"Failed to join group.",
		})

		session.Save(r, w)
		http.Redirect(w, r, "/groups/"+g.IDString(), http.StatusFound)
		return
	}

	session.AddFlash(framework.Flash{
		framework.FlashSuccess,
		"You're now a member of " + g.Name,
	})

	session.Save(r, w)
	http.Redirect(w, r, "/groups/"+g.IDString(), http.StatusFound)
	return
}
