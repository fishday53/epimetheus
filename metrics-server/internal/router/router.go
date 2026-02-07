package router

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/log"
	"metrics-server/internal/storage/memory"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexor() *chi.Mux {

	ctx_legacy := &handlers.AppContext{
		DB:  memory.NewMemStorage(),
		Log: log.NewLogger(),
	}

	ctx := &handlers.AppContext{
		DB:  memory.NewMemStorage(),
		Log: log.NewLogger(),
	}

	r := chi.NewRouter()

	// legacy plaintext API
	r.Group(func(r chi.Router) {
		r.Use(ctx_legacy.Logger)
		r.Get(`/value/{mtype}/{name}`, ctx_legacy.GetParam)
		r.Post(`/update/{mtype}/{name}/{value}`, ctx_legacy.SetParam)
	})

	// JSON API
	r.Group(func(r chi.Router) {
		r.Use(ctx.Logger)
		r.Use(ctx.CheckContentType)
		r.Post(`/value/`, ctx.GetParamJSON)
		r.Post(`/update/`, ctx.SetParamJSON)
		r.Get(`/`, ctx.GetAllParamsJSON)
	})

	return r
}
