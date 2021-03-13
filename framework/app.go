package framework

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
 * App represents the overall state of the application and what makes it unique,
 * particularly when it comes to environment variables. The database connection
 * and log level in particular and tracked here.
 */
type App struct {
	Router       *chi.Mux
	DB           *pgxpool.Pool
	hostname     string
	port         uint16
	ThemeName    string
	ThemeRoot    string
	Version      string
	name         string
	Slug         string
	LoggingLevel *slog.LevelVar
}

/*
 * Name returns the display name used by the App.
 */
func (a *App) Name() string {

	return a.name
}

/*
 * Port returns the TCP port that the app is listening on.
 */
func (a *App) Port() uint16 {

	return a.port
}

/*
 * Run starts the HTTP server up.
 */
func (a *App) Run() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", a.port), a.Router))
}

/*
 * NewApp returns a new instance of App with many of its fields initialized.
 * Zero or more AppOptions may be provided to customize default values.
 *
 * The called of this function should defer close the DB connection.
 */
func NewApp(appName, dbUser, dbPassword, dbHost string, dbPort int, dbName string, opts ...AppOption) (*App, error) {

	errs := Validator.Var(appName, "required")
	if errs != nil {
		return nil, fmt.Errorf("The app name cannot be blank.")
	}

	slog.Info("Initializing the web app.", "app", appName)

	// connect to database
	slog.Info("Attempting to connect to DB.", "connstr", fmt.Sprintf("postgres://%s:<redacted>@%s:%d/%s", dbUser, dbHost, dbPort, dbName))
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the database. Msg: %s", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("Database ping failed.")
	}

	// init struct and set some defaults
	a := &App{
		name:     appName,
		DB:       db,
		hostname: "localhost",
		port:     8000,
	}

	// apply options
	for _, opt := range opts {
		opt(a)
	}

	return a, nil
}
