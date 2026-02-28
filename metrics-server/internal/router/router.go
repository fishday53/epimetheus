package router

import (
	"metrics-server/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexer(app *handlers.AppContext) *chi.Mux {

	r := chi.NewRouter()

	r.Use(app.Logger)
	r.Use(app.GzipHandler)

	// legacy plaintext API
	r.Group(func(r chi.Router) {
		r.Get(`/value/{mtype}/{name}`, app.GetParam)
		r.Post(`/update/{mtype}/{name}/{value}`, app.SetParam)
		r.Get(`/`, app.GetAllParams)
		r.Get(`/ping`, app.CheckDBConnect)
	})

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(app.CheckContentType)
		r.Post(`/value/`, app.GetParamJSON)
		r.Post(`/update/`, app.SetParamJSON)
	})

	return r
}
