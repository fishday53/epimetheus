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

	r.Get(`/`, ctx.GetAllParams)
	r.Get(`/value/{kind}/{name}`, ctx.GetParam)
	r.Post(`/update/{kind}/{name}/{value}`, ctx.SetParam)

	return r
}
