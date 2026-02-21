package server

import (
	"metrics-server/internal/handlers"
	"metrics-server/internal/router"
	"metrics-server/internal/storage"
	"net/http"
	"time"

	"metrics-server/internal/config"
)

func Dumper(ctx *handlers.AppContext) {
	for {
		ctx.DB.Dump(ctx.Dump.Path)
		time.Sleep(time.Duration(ctx.Dump.Period) * time.Second)
	}
}

func HTTPServer() {

	var err error
	var cfg config.Config

	err = cfg.Get()
	if err != nil {
		panic(err)
	}

	ctx := handlers.NewAppContext("main", &storage.Dump{Path: cfg.FileStoragePath, Period: cfg.StoreInterval})

	if cfg.Restore {
		err := ctx.DB.Restore(cfg.FileStoragePath)
		if err != nil {
			panic(err)
		}
	}

	if cfg.StoreInterval > 0 {
		go Dumper(ctx)
	}

	err = http.ListenAndServe(cfg.Addr, router.NewMultiplexer(ctx))
	if err != nil {
		panic(err)
	}
}
