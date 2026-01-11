package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"time"
)

func main() {

	addr := flag.String("a", "localhost:8080", "Server address")
	reportInterval := flag.Int("r", 10, "Report interval")
	pollInterval := flag.Int("p", 2, "Poll interval")
	flag.Parse()

	url := "http://" + *addr + "/update"
	metrics := NewMetricsBatch()
	metrics.Counter["PollCount"] = 0

	for {
		for i := 0; i < (*reportInterval / *pollInterval); i++ {

			if err := metrics.getAllRuntimeMetrics(metricList); err != nil {
				fmt.Println(err)
			}

			metrics.Counter["PollCount"]++
			metrics.Gauge["RandomValue"] = rand.Float64()

			time.Sleep(time.Duration(*pollInterval) * time.Second)
		}

		if err := metrics.sendAllMetrics(url); err != nil {
			fmt.Println(err)
		}
	}
}
