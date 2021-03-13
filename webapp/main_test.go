package main

import (
	"log/slog"
	"os"
	"testing"

	"github.com/eventhunt-org/webapp/framework"

	"github.com/gorilla/sessions"
	"github.com/lmittmann/tint"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var a app
var userCreated = false

func TestMain(m *testing.M) {

	// Set defaults for config values
	viper.SetDefault("app_environment", "development")
	viper.SetDefault("app_scheme", "http")
	viper.SetDefault("app_host", "127.0.0.1")
	viper.SetDefault("app_port", 8200)
	viper.SetDefault("db_user", "app")
	viper.SetDefault("db_host", "127.0.0.1")
	viper.SetDefault("db_port", 8201)
	viper.SetDefault("db_name", "app")
	viper.SetDefault("auth_scheme", "http")
	viper.SetDefault("auth_host", "127.0.0.1")
	viper.SetDefault("auth_port", 8100)
	viper.SetDefault("auth_session_key", "CHANGE_ME")

	// Attempt to load config values from the `.env` file. If the file is not
	// found, that's okay.
	viper.SetConfigFile("../.env")
	viper.ReadInConfig()

	// Attempt to load config values from environment variables. Most useful in
	// non development environments.
	viper.AutomaticEnv()

	/*
	 * Setup Logging. The style of log output will vary depending on the
	 * environment we're running in. In dev we're focused more on human
	 * legibility and simplicity. The log level also varies.
	 */
	switch viper.GetString("environment") {
	case "production":
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				TimeFormat: "2006/01/02 15:04:05",
			}),
		))
	case "staging":
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: "2006/01/02 15:04:05",
			}),
		))
	default: // also development
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: "15:04:05",
			}),
		))
	}

	log.Info("Starting app...")

	store = sessions.NewCookieStore([]byte(viper.GetString("auth_session_key")))

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

	a = app{innerApp}

	a.Initialize(
		os.Getenv("APP_THEME_ROOT"),
		"original",
	)

	environment = "testing"

	slog.Info("App initialized.", "mode", environment)

	exitVal := m.Run()

	os.Exit(exitVal)
}
