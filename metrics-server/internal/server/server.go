package server

import (
	"metrics-server/internal/router"
	"net/http"

	"metrics-server/internal/config"
)

func HTTPServer() {

	var cfg config.Config
	cfg.Get()

	err := http.ListenAndServe(cfg.Addr, router.NewMultiplexor())
	if err != nil {
		panic(err)
	}
}
