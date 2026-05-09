// Package agent is used to gather all kinds of metrics and send them via http.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"metrics-agent/internal/config"
	"metrics-agent/internal/crypt"
	"metrics-agent/internal/metrics"
	pb "metrics-agent/internal/proto"
	"metrics-agent/internal/ratelimit"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	tick int64 = 1

	backoffSchedule = []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
)

type Transport interface {
	SendMetrics(metric *metrics.Batch) error
}

type HTTPTransport struct {
	endpoint string
	client   *http.Client
	hashKey  string
	pubKey   *rsa.PublicKey
}

type GRPCTransport struct {
	conn   *grpc.ClientConn
	client pb.MetricServiceClient
}

// NewTransport is a Transport construcror
func NewTransport(cfg *config.Config, url string, pubKey *rsa.PublicKey) (Transport, error) {
	switch cfg.Transport {
	case "http":
		return &HTTPTransport{
			endpoint: url,
			client:   &http.Client{},
			hashKey:  cfg.HashKey,
			pubKey:   pubKey,
		}, nil
	case "grpc":
		conn, err := grpc.Dial(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		return &GRPCTransport{
			conn:   conn,
			client: pb.NewMetricServiceClient(conn),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", cfg.Transport)
	}
}

// SendMetrics sends a signed batch of metrics via http.
func (h *HTTPTransport) SendMetrics(metric *metrics.Batch) error {
	var hashHeader string

	jsonData, err := json.Marshal(metric)

	if err != nil {
		return fmt.Errorf("error in marshaller: %v", err)
	}

	if h.hashKey != "" {
		hashHeader = getHash(h.hashKey, jsonData)
	}

	if h.pubKey != nil {
		jsonData, err = crypt.Encrypt(h.pubKey, jsonData)
		if err != nil {
			return fmt.Errorf("cannot encrypt jsonData: %v", err)
		}
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
		req, err := http.NewRequest("POST", h.endpoint, &buf)
		if err != nil {
			log.Printf("Error creating http-request: %v\n", err)
			time.Sleep(backoff)
			continue
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Hashsha256", hashHeader)

		xRealIP, err := getLocalIP()
		if err != nil {
			return fmt.Errorf("cannot set X-Real-IP header: %v", err)
		}
		req.Header.Set("X-Real-IP", xRealIP)

		resp, err := h.client.Do(req)
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

// SendMetrics sends a signed batch of metrics via grpc.
func (g *GRPCTransport) SendMetrics(metric *metrics.Batch) error {
	xRealIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("cannot set X-Real-IP header: %v", err)
	}
	md := metadata.New(map[string]string{
		"x-real-ip": xRealIP,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	pbBatch := convertToProtoBatch(metric)

	for _, backoff := range backoffSchedule {
		resp, err := g.client.SetMultiParamGRPC(ctx, &pb.AddMetricsRequest{
			Metrics: pbBatch,
		})
		if err != nil {
			log.Printf("Error in grpc-connect: %v\n", err)
			time.Sleep(backoff)
			continue
		} else {
			if resp.Error != "" {
				log.Printf("Error in grpc-request: %v\n", resp.Error)
				time.Sleep(backoff)
				continue
			} else {
				break
			}
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
	stopWork <-chan struct{},
	limit *ratelimit.TokenBucketLimiter,
) {
	var (
		pubKey *rsa.PublicKey
		err    error
	)

	ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer ticker.Stop()

	defer wg.Done()

	if cfg.CryptoKeyPath != "" {
		pubKey, err = crypt.GetPublicKey(cfg.CryptoKeyPath)
		if err != nil {
			log.Fatalf("crypto key %s get failed:%v\n", cfg.CryptoKeyPath, err)
			return
		}
	}

	transport, err := NewTransport(cfg, url, pubKey)
	if err != nil {
		log.Fatalf("%s transport initialization failed:%v\n", cfg.Transport, err)
		return
	}

	for range ticker.C {
	SendLoop:
		for {
			for limit.Allow() {
				select {
				case j := <-jobs:
					err := transport.SendMetrics(j)
					if err != nil {
						log.Printf("Metric send failed. Error:%v\n", err)
					}
				case <-stopWork:
					log.Printf("Agent Shutdown gracefully")
					return
				default:
					break SendLoop
				}
			}
		}
	}
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("cannot list network interfaces: %v", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("cannot find local IP address")
}

func convertToProtoBatch(batch *metrics.Batch) *pb.Batch {
	if batch == nil {
		return nil
	}

	pbBatch := &pb.Batch{
		Metrics: make([]*pb.Metric, 0, len(*batch)),
	}

	for _, m := range *batch {
		pbMetric := &pb.Metric{
			ID:    m.ID,
			MType: m.MType,
			Delta: *m.Delta,
			Value: *m.Value,
		}
		pbBatch.Metrics = append(pbBatch.Metrics, pbMetric)
	}

	return pbBatch
}
