package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func main() {

	url := "http://localhost:8080/update"
	pollInterval := 2
	reportInterval := 10
	metrics := MetricsBatch{}
	metrics.Gauge = make(map[string]gauge)
	metrics.Counter = make(map[string]counter)
	metrics.Counter["PollCount"] = 0

	for {
		for i := 0; i < (reportInterval / pollInterval); i++ {

			if err := metrics.getAllRuntimeMetrics(metricList); err != nil {
				fmt.Println(err)
			}

			metrics.Counter["PollCount"]++
			metrics.Gauge["RandomValue"] = gauge(rand.Float64())

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}

		if err := metrics.sendAllMetrics(url); err != nil {
			fmt.Println(err)
		}
	}
}
