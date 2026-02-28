package main

import (
	"log"
	"metrics-server/internal/config"
	"metrics-server/internal/handlers"
	"metrics-server/internal/server"
	"os"
	"time"
)

func Dumper(app *handlers.AppContext) {
	for {
		app.DB.Dump(app.Cfg.FileStoragePath)
		time.Sleep(time.Duration(app.Cfg.StoreInterval) * time.Second)
	}
}

func main() {

	var err error
	var cfg config.Config

	log.SetOutput(os.Stdout)

	err = cfg.Get()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	app, err := handlers.NewAppContext(&cfg)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	if cfg.Restore {
		err := app.DB.Restore(cfg.FileStoragePath)
		if err != nil {
			app.Log.Fatalf("%v", err)
			return
		}
	}

	if cfg.StoreInterval > 0 {
		go Dumper(app)
	}

	server.HTTPServer(app)
}
