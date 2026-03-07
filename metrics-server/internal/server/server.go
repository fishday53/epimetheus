package server

import (
	"metrics-server/internal/router"
	"metrics-server/internal/usecase/context"
	"net/http"
)

func HTTPServer(app *context.AppContext) {
	err := http.ListenAndServe(app.Cfg.Addr, router.NewMultiplexer(app))
	if err != nil {
		app.Log.Fatalf("%v", err)
		return
	}
}
