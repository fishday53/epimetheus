package main

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
			list:    metricList, // []string{"Frees"},
			wantErr: false,
		},
		{
			name:    "Nonexistent metric",
			list:    []string{"JustRandonString"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetricsBatch{}
			m.Gauge = make(map[string]gauge)
			gotErr := m.getAllRuntimeMetrics(tt.list)
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
			m := MetricsBatch{}
			m.Gauge = make(map[string]gauge)
			m.Counter = make(map[string]counter)
			m.Gauge["RandomValue"] = gauge(rand.Float64())
			m.Counter["PollCount"] = 0

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			gotErr := m.sendAllMetrics(server.URL)
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
