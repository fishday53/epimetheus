package server

import (
	"log"
	"metrics-server/internal/handlers"
	"metrics-server/internal/router"
	"net/http"
	"os"
	"time"

	"metrics-server/internal/config"
)

func Dumper(ctx *handlers.AppContext) {
	for {
		ctx.DB.Dump(ctx.Cfg.FileStoragePath)
		time.Sleep(time.Duration(ctx.Cfg.StoreInterval) * time.Second)
	}
}

func HTTPServer() {

	var err error
	var cfg config.Config

	log.SetOutput(os.Stdout)

	err = cfg.Get()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	ctx, err := handlers.NewAppContext(&cfg)
	if err != nil {
		log.Printf("%v", err)
		return
	}

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
