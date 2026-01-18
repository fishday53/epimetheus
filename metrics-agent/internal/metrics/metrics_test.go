package metrics

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsBatch_getAllRuntimeMetrics(t *testing.T) {
	tests := []struct {
		name    string
		list    []string
		wantErr bool
	}{
		{
			name:    "Existed metrics",
			list:    MetricList, // []string{"Frees"},
			wantErr: false,
		},
		{
			name:    "Nonexistent metric",
			list:    []string{"JustRandomString"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetricsBatch()
			gotErr := m.GetAllRuntimeMetrics(tt.list)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getAllRuntimeMetrics() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getAllRuntimeMetrics() succeeded unexpectedly")
			}
		})
	}
}

func TestMetricsBatch_sendAllMetrics(t *testing.T) {
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
			m := NewMetricsBatch()
			m.Gauge["RandomValue"] = rand.Float64()
			m.Counter["PollCount"] = 0

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			gotErr := m.SendAllMetrics(server.URL)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("sendAllMetrics() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("sendAllMetrics() succeeded unexpectedly")
			}
		})
	}
}
