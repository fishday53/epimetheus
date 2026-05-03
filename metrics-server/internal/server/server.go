// Package server implements an http-server.
package server

import (
	ctx "context"
	"log"
	"metrics-server/internal/router"
	"metrics-server/internal/usecase/context"
	"net/http"
)

// HTTPServer starts to serve http-requests.
func HTTPServer(app *context.AppContext, idleConnsClosed chan struct{}) {

	var srv = http.Server{Addr: app.Cfg.Addr, Handler: router.NewMultiplexer(app)}

	go func() {
		<-app.Stop
		if err := srv.Shutdown(ctx.Background()); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		app.Log.Fatalf("%v", err)
		return
	}
}
