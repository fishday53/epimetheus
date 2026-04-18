package main

import (
	"fmt"
	"log"
	"metrics-server/internal/config"
	"metrics-server/internal/server"
	"metrics-server/internal/storage"
	"metrics-server/internal/usecase/context"
	"os"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		printVal(buildVersion), printVal(buildDate), printVal(buildCommit),
	)

	var (
		err error
		cfg config.Config
	)

	log.SetOutput(os.Stdout)

	err = cfg.Get()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	app, err := context.NewAppContext(&cfg)
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
		go storage.Dumper(app)
	}

	server.HTTPServer(app)
}

func printVal(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
