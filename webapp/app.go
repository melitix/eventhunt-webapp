package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

const (
	HostnameBase  = "melitix.events"
	HostnameEmail = "mail.melitix.com"
)

// Making this global just because I kind of need to. I would like this to be
// ap specific but doesn't work for access from models
// Update: I'm considering moving this to the new db package.
var validate *validator.Validate

type app struct {
	*framework.App
}

func (a *app) Initialize(themeRoot, themeName string) {

	if themeRoot == "" {
		themeRoot = "./"
	}

	a.ThemeName = themeName
	a.ThemeRoot = themeRoot

	if version != "dev" {
		a.Version = "v" + version
	} else {
		a.Version = version
	}

	validate = validator.New()

	setupMenus()

	a.Router = chi.NewRouter()

	// setup middleware
	a.Router.Use(middleware.RedirectSlashes)
	a.Router.Use(a.loggingMiddleware)

	a.Router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", http.FileServer(http.Dir(a.ThemePath()+"assets"))).ServeHTTP(w, r)
	})
	a.initializeRoutes()

	a.LoggingLevel = new(slog.LevelVar)
}

// Builds the theme PATH and then returns it
func (a *app) ThemePath() string {
	return a.ThemeRoot + "themes/" + a.ThemeName + "/"
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Error(message)
}

/*
 * Return a JSON payload for API requests.
 */
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

/*
 * Inits the app and underlaying framework.App.
 */
func NewApp() {
}
