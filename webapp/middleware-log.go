package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

/*
 * Middleware - logs incoming HTTP requests. Measures how long the request took
 * to serve.
 */
func (a *app) loggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// don't log css/images/JS assets
		if strings.HasPrefix(r.URL.Path, "/assets") || strings.HasPrefix(r.URL.Path, "/static") || strings.HasPrefix(r.URL.Path, "/favicon.") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info(fmt.Sprintf("[%s] %s", r.Method, r.RequestURI), "duration", time.Since(start).Round(time.Millisecond))

	})
}
