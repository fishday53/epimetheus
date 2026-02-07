package router

import (
	"metrics-server/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexor() *chi.Mux {

	ctxLegacy := handlers.NewAppContext("legacy")
	ctxJSON := handlers.NewAppContext("json")

	r := chi.NewRouter()

	// legacy plaintext API
	r.Group(func(r chi.Router) {
		r.Use(ctxLegacy.Logger)
		r.Get(`/value/{mtype}/{name}`, ctxLegacy.GetParam)
		r.Post(`/update/{mtype}/{name}/{value}`, ctxLegacy.SetParam)
	})

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(ctxJSON.Logger)
		r.Use(ctxJSON.CheckContentType)
		r.Post(`/value/`, ctxJSON.GetParamJSON)
		r.Post(`/update/`, ctxJSON.SetParamJSON)
		r.Get(`/`, ctxJSON.GetAllParamsJSON)
	})

	return r
}
