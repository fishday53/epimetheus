package router

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/usecase/context"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexer(app *context.AppContext) *chi.Mux {

	r := chi.NewRouter()

	r.Use(handlers.Logger(app))
	//r.Use(handlers.HashHandler(app))
	r.Use(handlers.GzipHandler(app))

	// legacy plaintext API
	r.Group(func(r chi.Router) {
		r.Get(`/value/{mtype}/{name}`, handlers.GetParam(app))
		r.Post(`/update/{mtype}/{name}/{value}`, handlers.SetParam(app))
		r.Get(`/`, handlers.GetAllParams(app))
		r.Get(`/ping`, handlers.CheckDBConnect(app))
	})

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(handlers.CheckContentType(app))
		r.Post(`/value/`, handlers.GetParamJSON(app))
		r.Post(`/update/`, handlers.SetParamJSON(app))
		r.Post(`/updates/`, handlers.SetMultiParamJSON(app))
	})

	return r
}
