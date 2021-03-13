package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/eventhunt-org/webapp/framework"
	"github.com/eventhunt-org/webapp/webapp/nonce"
	"github.com/gorilla/sessions"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// default global values
var (
	AppName     = "EventHunt"
	version     = "dev"
	environment = "development"
	store       *sessions.CookieStore
	ns          *nonce.Store
	hostname    = "127.0.0.1"
)

func main() {

	// Set defaults for config values
	viper.SetDefault("app_environment", "development")
	viper.SetDefault("app_scheme", "http")
	viper.SetDefault("app_host", "127.0.0.1")
	viper.SetDefault("app_port", 9000)

	viper.SetDefault("db_user", "app")
	viper.SetDefault("db_host", "127.0.0.1")
	viper.SetDefault("db_port", 9001)
	viper.SetDefault("db_name", "app")

	viper.SetDefault("auth_scheme", "http")
	viper.SetDefault("auth_host", "127.0.0.1")
	viper.SetDefault("auth_port", 9100)
	viper.SetDefault("auth_session_key", "CHANGE_ME")

	// Attempt to load config values from the `.env` file. If the file is not
	// found, that's okay.
	viper.SetConfigFile("../.env")
	viper.ReadInConfig()

	// Attempt to load config values from environment variables. Most useful in
	// non development environments.
	viper.AutomaticEnv()

	// setup app global variables
	store = sessions.NewCookieStore([]byte(viper.GetString("auth_session_key")))

	/*
	 * Setup Logging. The style of log output will vary depending on the
	 * environment we're running in. In dev we're focused more on human
	 * legibility and simplicity. The log level also varies.
	 */
	switch viper.GetString("app_environment") {
	case "production":
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				TimeFormat: "2006/01/02 15:04:05",
			}),
		))
		store.Options.Domain = "melitix.com"
	case "staging":
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: "2006/01/02 15:04:05",
			}),
		))
		store.Options.Domain = "staging.melitix.com"
	default: // also development
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: "15:04:05",
			}),
		))
		store.Options.Domain = "127.0.0.1"
	}

	slog.Info("Starting app...")

	var err error
	ns, err = nonce.New()
	if err != nil {
		log.Fatal("NonceService failed to initialize.")
	}

	slog.Info("Using a DB hostname.", "db_host", viper.GetString("db_host"))

	innerApp, err := framework.NewApp(
		AppName,
		viper.GetString("db_user"),
		viper.GetString("db_pass"),
		viper.GetString("db_host"),
		viper.GetInt("db_port"),
		viper.GetString("db_name"),
		framework.WithPort(viper.GetUint16("app_port")),
	)
	if err != nil {
		slog.Error(err.Error())
		log.Fatal("Creating inner app failed.")
	}

	a := app{innerApp}

	a.Initialize(
		os.Getenv("APP_THEME_ROOT"),
		"original",
	)

	slog.Info("App initialized.", "mode", environment)
	slog.Info(fmt.Sprintf("The webapp can be viewed at http://%s:%d", viper.GetString("app_host"), viper.GetUint16("app_port")))

	a.Run()
}
