// Package server implements an http-server.
package server

import (
	"metrics-server/internal/router"
	"metrics-server/internal/usecase/context"
	"net/http"
)

// HTTPServer starts to serve http-requests.
func HTTPServer(app *context.AppContext) {
	err := http.ListenAndServe(app.Cfg.Addr, router.NewMultiplexer(app))
	if err != nil {
		app.Log.Fatalf("%v", err)
		return
	}
}
