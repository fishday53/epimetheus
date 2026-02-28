package main

import (
	"log"
	"metrics-agent/internal/agent"
	"metrics-agent/internal/config"
	"os"
	"time"
)

func main() {

	var cfg config.Config

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

	for {
		for i := 0; i < (cfg.ReportInterval / cfg.PollInterval); i++ {

			m, err := agent.GetMetrics(&cfg)
			if err != nil {
				log.Printf("Cannot get metrics: %v\n", err)
			}

			if len(*m) != 0 {
				err = agent.SendMetrics(url, m)
				if err != nil {
					log.Printf("Metric send failed. Error:%v\n", err)
				}
			}

			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}
}
