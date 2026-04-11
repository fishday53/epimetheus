package main

import (
	"context"
	"log"
	"metrics-agent/internal/agent"
	"metrics-agent/internal/config"
	"metrics-agent/internal/ratelimit"
	"os"
	"sync"
	"time"
)

func main() {

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
