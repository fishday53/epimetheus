package server

import (
	"log"
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
		log.Fatalf("%v", err)
		return
	}

	ctx := handlers.NewAppContext("main", &storage.Dump{Path: cfg.FileStoragePath, Period: cfg.StoreInterval})
	defer ctx.Log.Sync()

	if cfg.Restore {
		err := ctx.DB.Restore(cfg.FileStoragePath)
		if err != nil {
			ctx.Log.Fatalf("%v", err)
			return
		}
	}

	if cfg.StoreInterval > 0 {
		go Dumper(ctx)
	}

	err = http.ListenAndServe(cfg.Addr, router.NewMultiplexer(ctx))
	if err != nil {
		ctx.Log.Fatalf("%v", err)
		return
	}
}
