package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"metrics-agent/internal/config"
	"metrics-agent/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_SendMetric(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Check sending",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rnd := rand.Float64()
			randomValue := metrics.Metric{
				ID:    "RandomValue",
				MType: "gauge",
				Value: &rnd,
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			gotErr := SendMetrics(server.URL, "", &metrics.Batch{&randomValue})
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("sendMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("sendMetric() succeeded unexpectedly")
			}
		})
	}
}

func Test_GetMetricsRuntime(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Check GetMetricsRuntime",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{BufferSize: 1, PollInterval: 1}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			ch1 := GetMetricsRuntime(ctx, &cfg)

			batch := <-ch1
			for _, b := range *batch {
				if b.Delta == nil && b.Value == nil {
					t.Errorf("Can't get %s metric", b.ID)
				}
			}
			cancel()
		})
	}
}

func Test_GetMetricsVMstat(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Check GetMetricsVMstat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{BufferSize: 1, PollInterval: 1}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			ch2 := GetMetricsVMstat(ctx, &cfg)

			batch := <-ch2
			for _, b := range *batch {
				if b.Delta == nil && b.Value == nil {
					t.Errorf("Can't get %s metric", b.ID)
				}
			}
			cancel()
		})
	}
}
