package main

import (
	"fmt"
	"log"
	"metrics-server/internal/config"
	"metrics-server/internal/server"
	"metrics-server/internal/storage"
	"metrics-server/internal/usecase/context"
	"os"
	"os/signal"
	"syscall"
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

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigint
		close(app.Stop)
	}()

	server.HTTPServer(app, idleConnsClosed)

	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")
}

func printVal(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
