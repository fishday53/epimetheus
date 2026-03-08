package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func GetMetrics(cfg *config.Config) (*[]*metrics.Metric, error) {

	var m []*metrics.Metric

	log.SetOutput(os.Stdout)

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
	m = append(m, &pollCount)

	// Additional gauge
	rnd := rand.Float64()
	randomValue := metrics.Metric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &rnd,
	}
	m = append(m, &randomValue)

	return &m, nil
}

func SendMetrics(url, hashKey string, metric *[]*metrics.Metric) error {

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
