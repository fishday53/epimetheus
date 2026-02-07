package router

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/log"
	"metrics-server/internal/storage/memory"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexor() *chi.Mux {

	ctx := &handlers.AppContext{
		DB:  memory.NewMemStorage(),
		Log: log.NewLogger(),
	}

	r := chi.NewRouter()

	r.Use(ctx.Logger)

	r.Get(`/`, ctx.GetAllParamsJSON)

	// legacy plaintext API
	r.Get(`/value/{mtype}/{name}`, ctx.GetParam)
	r.Post(`/update/{mtype}/{name}/{value}`, ctx.SetParam)

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(ctx.CheckContentType)
		r.Post(`/value/`, ctx.GetParamJSON)
		r.Post(`/update/`, ctx.SetParamJSON)
	})

	return r
}
