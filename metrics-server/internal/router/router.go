package router

import (
	"metrics-server/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexor(ctx *handlers.AppContext) *chi.Mux {

	//ctx := handlers.NewAppContext("main")

	r := chi.NewRouter()

	r.Use(ctx.Logger)
	r.Use(ctx.GzipHandler)

	// legacy plaintext API
	r.Group(func(r chi.Router) {
		r.Get(`/value/{mtype}/{name}`, ctx.GetParam)
		r.Post(`/update/{mtype}/{name}/{value}`, ctx.SetParam)
		r.Get(`/`, ctx.GetAllParams)
	})

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(ctx.CheckContentType)
		r.Post(`/value/`, ctx.GetParamJSON)
		r.Post(`/update/`, ctx.SetParamJSON)
		//r.Get(`/`, ctx.GetAllParamsJSON)
	})

	return r
}
