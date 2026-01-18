package server

import (
	"metrics-server/internal/router"
	"net/http"

	"metrics-server/internal/config"
)

func HTTPServer() {

	var err error
	var cfg config.Config

	err = cfg.Get()
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(cfg.Addr, router.NewMultiplexor())
	if err != nil {
		panic(err)
	}
}
