package main

import (
	"log"
	"metrics-agent/internal/agent"
	"metrics-agent/internal/config"
	"os"
	"sync"
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
		proto  = "http://"
		path   = "/updates/"
		buffer = 100
	)

	url := proto + cfg.Addr + path
	cfg.BufferSize = buffer

	ch1 := agent.GetMetrics1(&cfg)
	ch2 := agent.GetMetrics15(&cfg)

	for w := 1; w <= cfg.RateLimit; w++ {
		wg.Add(1)
		go agent.SendWorker(&wg, &cfg, url, agent.FanIn(ch1, ch2))
	}

	wg.Wait()
}
