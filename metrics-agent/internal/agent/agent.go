// Package agent is used to gather all kinds of metrics and send them via http.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"metrics-agent/internal/config"
	"metrics-agent/internal/metrics"
	"metrics-agent/internal/ratelimit"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	tick int64 = 1

	backoffSchedule = []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
)

// SendMetrics sends a signed batch of metrics via http.
func SendMetrics(url, hashKey string, metric *metrics.Batch) error {
	var hashHeader string

	jsonData, err := json.Marshal(metric)

	if err != nil {
		return fmt.Errorf("error in marshaller: %v", err)
	}

	if hashKey != "" {
		hashHeader = getHash(hashKey, jsonData)
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
		req.Header.Set("Hashsha256", hashHeader)

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

func getHash(hashKey string, b []byte) string {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(b[:])
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes[:])
}

// GetMetricsRuntime gathers runtime metrics.
func GetMetricsRuntime(ctx context.Context, cfg *config.Config) chan *metrics.Batch {
	outChan := make(chan *metrics.Batch, cfg.BufferSize)
	ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)

	go func() {
		defer close(outChan)
		defer ticker.Stop()

		log.SetOutput(os.Stdout)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var m metrics.Batch
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

					m = append(m, &metric)
				}

				// Additional counter
				pollCount := metrics.Metric{
					ID:    "PollCount",
					MType: "counter",
					Delta: &tick,
				}
				log.Printf("PollCount=%d\n", tick)
				m = append(m, &pollCount)

				// Additional gauge
				rnd := rand.Float64()
				randomValue := metrics.Metric{
					ID:    "RandomValue",
					MType: "gauge",
					Value: &rnd,
				}
				log.Printf("RandomValue=%f\n", rnd)
				m = append(m, &randomValue)

				outChan <- &m
			}
		}
	}()

	return outChan
}

// GetMetricsVMstat gathers VMstat metrics.
func GetMetricsVMstat(ctx context.Context, cfg *config.Config) chan *metrics.Batch {
	outChan := make(chan *metrics.Batch, cfg.BufferSize)
	ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)

	go func() {
		defer close(outChan)
		defer ticker.Stop()

		log.SetOutput(os.Stdout)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var m metrics.Batch
				// VM metrics from PS
				for _, metricName := range metrics.VMMetrics {

					value, err := metrics.GetVMStatMetric(metricName)
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

					m = append(m, &metric)
				}

				// CPU utilizarion from PS
				cpuUtil, err := metrics.GetCPUTotal()
				if err != nil {
					log.Printf("CPUutilization1 error: %v\n", err)
				} else {
					log.Printf("CPUutilization1=%f\n", cpuUtil)
				}
				cpuUtilValue := metrics.Metric{
					ID:    "CPUutilization1",
					MType: "gauge",
					Value: &cpuUtil,
				}
				log.Printf("CPUutilization1=%f\n", cpuUtil)
				m = append(m, &cpuUtilValue)
				outChan <- &m
			}
		}
	}()

	return outChan
}

// FanIn joins and publishes all metrics kinds in a single queue to process.
func FanIn(chs ...chan *metrics.Batch) chan *metrics.Batch {
	finalCh := make(chan *metrics.Batch)

	var wg sync.WaitGroup

	for _, ch := range chs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				finalCh <- data
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

// SendWorker is used to process the metrics queue.
func SendWorker(
	wg *sync.WaitGroup,
	cfg *config.Config,
	url string,
	jobs <-chan *metrics.Batch,
	limit *ratelimit.TokenBucketLimiter,
) {
	ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer ticker.Stop()

	defer wg.Done()

	for range ticker.C {
	SendLoop:
		for {
			for limit.Allow() {
				select {
				case j := <-jobs:
					err := SendMetrics(url, cfg.HashKey, j)
					if err != nil {
						log.Printf("Metric send failed. Error:%v\n", err)
					}
				default:
					break SendLoop
				}
			}
		}
	}
}
