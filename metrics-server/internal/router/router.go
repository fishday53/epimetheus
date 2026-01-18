package router

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/storage/memory"

	"github.com/go-chi/chi/v5"
)

func NewMultiplexor() *chi.Mux {

	ctx := &handlers.AppContext{DB: memory.NewMemStorage()}

	r := chi.NewRouter()
	r.Get(`/`, ctx.GetAllParams)
	r.Get(`/value/{kind}/{name}`, ctx.GetParam)
	r.Post(`/update/{kind}/{name}/{value}`, ctx.SetParam)

	return r
}
