package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"metrics-agent/internal/config"
	"metrics-agent/internal/metrics"
	"net/http"
	"os"
	"time"
)

var tick int64 = 1

var backoffSchedule = []time.Duration{
	100 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
}

func sendMetric(url string, metric *metrics.Metric) error {

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("error in marshaller: %v", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(jsonData); err != nil {
		return fmt.Errorf("error gzipping data: %v", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("error closing gzip writer: %v", err)
	}

	for _, backoff := range backoffSchedule {
		req, err := http.NewRequest("POST", url, &buf)
		if err != nil {
			log.Printf("Error creating http-request: %v\n", err)
			time.Sleep(backoff)
			continue
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error posting query: %v\n", err)
			time.Sleep(backoff)
		} else {
			defer resp.Body.Close()
			break
		}
	}

	return nil
}

func Agent() error {
	var cfg config.Config

	log.SetOutput(os.Stdout)

	if err := cfg.Get(); err != nil {
		log.Printf("Cannot get configuration. Error:%v\n", err)
		return err
	}

	url := "http://" + cfg.Addr + "/update/"

	for {
		for i := 0; i < (cfg.ReportInterval / cfg.PollInterval); i++ {

			// RunTime metrics
			for _, metricName := range metrics.MetricList {

				value, err := metrics.GetRuntimeMetric(metricName)
				if err != nil {
					log.Printf("%s error: %v\n", metricName, err)
				} else {
					log.Printf("%s=%f\n", metricName, value)
				}

				metric := metrics.Metric{
					ID:    metricName,
					MType: "gauge",
					Value: &value,
				}

				err = sendMetric(url, &metric)
				if err != nil {
					log.Fatalf("Metric send failed. Metric: %s, Error:%v\n", metricName, err)
					return err
				}
			}

			// Additional counter
			pollCount := metrics.Metric{
				ID:    "PollCount",
				MType: "counter",
				Delta: &tick,
			}
			err := sendMetric(url, &pollCount)
			if err != nil {
				log.Fatalf("Metric send failed. Metric: PollCount, Error:%v\n", err)
				return err
			}

			// Additional gauge
			rnd := rand.Float64()
			randomValue := metrics.Metric{
				ID:    "RandomValue",
				MType: "gauge",
				Value: &rnd,
			}
			err = sendMetric(url, &randomValue)
			if err != nil {
				log.Fatalf("Metric send failed. Metric: RandomValue, Error:%v\n", err)
				return err
			}

			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}

	}
}
