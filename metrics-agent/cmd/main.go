package main

import (
	"context"
	"fmt"
	"log"
	"metrics-agent/internal/agent"
	"metrics-agent/internal/config"
	"metrics-agent/internal/ratelimit"
	"os"
	"sync"
	"time"
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
		wg  sync.WaitGroup
		cfg config.Config
	)

	log.SetOutput(os.Stdout)

	if err := cfg.Get(); err != nil {
		log.Printf("Cannot get configuration. Error:%v\n", err)
		return
	}

	const (
		proto = "http://"
		path  = "/updates/"
	)

	url := proto + cfg.Addr + path
	cfg.BufferSize = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := agent.GetMetricsRuntime(ctx, &cfg)
	ch2 := agent.GetMetricsVMstat(ctx, &cfg)

	rateLimit := ratelimit.NewTokenBucketLimiter(ctx, cfg.RateLimit, time.Second*1)

	wg.Add(1)
	go agent.SendWorker(&wg, &cfg, url, agent.FanIn(ch1, ch2), rateLimit)

	wg.Wait()
}

func printVal(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
