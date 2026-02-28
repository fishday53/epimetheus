package server

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/router"
	"net/http"
)

func HTTPServer(app *handlers.AppContext) {
	err := http.ListenAndServe(app.Cfg.Addr, router.NewMultiplexer(app))
	if err != nil {
		app.Log.Fatalf("%v", err)
		return
	}
}
