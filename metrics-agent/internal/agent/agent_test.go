package agent

import (
	"fmt"
	"math/rand/v2"
	"metrics-agent/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
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

			gotErr := SendMetrics(server.URL, &[]*metrics.Metric{&randomValue})
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
