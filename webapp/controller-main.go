package main

import (
	"log/slog"
	"net/http"

	"github.com/eventhunt-org/webapp/webapp/db"

	"github.com/spf13/viper"
)

/*
 * Display / GET
 */
func (a *app) homepage(w http.ResponseWriter, r *http.Request) {

	// we may or may not have a logged in user
	u, ok := r.Context().Value("user").(*db.User)

	mapKey := viper.GetString("app_map_key")

	events, err := db.GetEvents(a.DB, 100)
	if err != nil {
		slog.Error(err.Error())
	}

	var groups []*db.Group
	if ok {
		groups, err = db.GetGroupsByUser(u)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	renderPage(a, "homepage", w, r, map[string]interface{}{
		"User":   u,
		"MapKey": mapKey,
		"Events": events,
		"Groups": groups,
	})
}
