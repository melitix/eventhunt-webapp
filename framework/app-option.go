package framework

import "log/slog"

/*
 * AppOption customizes how an App is configured via the constructor.
 */
type AppOption func(*App)

/*
 * WithPort returns an AppOption to set a custom HTTP port for an App. Logs a
 * warning when a privileged port is specified.
 */
func WithPort(port uint16) AppOption {
	return func(a *App) {

		a.port = port

		if a.port < 1024 {
			slog.Warn("The app port was set to a privileged port. If this app isn't run with the appropriate permissions, the server will fail to start.", "port", a.port)
		}
	}
}
