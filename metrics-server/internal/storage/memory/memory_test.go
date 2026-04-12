package memory

import (
	"metrics-server/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCounter   int64   = 527
	testGauge     float64 = 0.00005
	resultCounter int64   = testCounter + testCounter
	resultGauge   float64 = testGauge
)

func Test_Set(t *testing.T) {
	tests := []struct {
		name    string
		metric  usecase.Metric
		result  usecase.Metric
		wantErr bool
	}{
		{
			name:    "Set counter",
			metric:  usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter},
			result:  usecase.Metric{ID: "c1", MType: "counter", Delta: &resultCounter},
			wantErr: false,
		},
		{
			name:    "Set gauge",
			metric:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			result:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name:    "Change type for counter",
			metric:  usecase.Metric{ID: "c1", MType: "gauge", Value: &testGauge},
			wantErr: true,
		},
		{
			name:    "Change type for gauge",
			metric:  usecase.Metric{ID: "g1", MType: "counter", Delta: &testCounter},
			wantErr: true,
		},
		{
			name:    "Unsupported type",
			metric:  usecase.Metric{ID: "x1", MType: "unsupported", Delta: &testCounter},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				Metrics: map[string]MetricParam{
					"c1": {
						MType: "counter",
						Delta: &testCounter,
					},
					"g1": {
						MType: "gauge",
						Value: &testGauge,
					},
				},
			}
			result, err := storage.Set(&tt.metric)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("Set() succeeded unexpectedly")
				return
			}
			assert.Equal(t, result, &tt.result)
		})
	}
}

func Test_Get(t *testing.T) {
	tests := []struct {
		name    string
		metric  usecase.Metric
		result  usecase.Metric
		wantErr bool
	}{
		{
			name:    "Get counter",
			metric:  usecase.Metric{ID: "c1", MType: "counter"},
			result:  usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter},
			wantErr: false,
		},
		{
			name:    "Get gauge",
			metric:  usecase.Metric{ID: "g1", MType: "gauge"},
			result:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name:    "Wrong type for counter",
			metric:  usecase.Metric{ID: "c1", MType: "gauge"},
			wantErr: true,
		},
		{
			name:    "Wrong type for gauge",
			metric:  usecase.Metric{ID: "g1", MType: "counter"},
			wantErr: true,
		},
		{
			name:    "Unsupported type",
			metric:  usecase.Metric{ID: "x1", MType: "unsupported"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				Metrics: map[string]MetricParam{
					"c1": {
						MType: "counter",
						Delta: &testCounter,
					},
					"g1": {
						MType: "gauge",
						Value: &testGauge,
					},
				},
			}
			result, err := storage.Get(&tt.metric)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("Get() succeeded unexpectedly")
				return
			}
			assert.Equal(t, result, &tt.result)
		})
	}
}

func Test_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		result []usecase.Metric
	}{
		{
			name: "Get all",
			result: []usecase.Metric{
				{ID: "c1", MType: "counter", Delta: &testCounter},
				{ID: "g1", MType: "gauge", Value: &testGauge},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				Metrics: map[string]MetricParam{
					"c1": {
						MType: "counter",
						Delta: &testCounter,
					},
					"g1": {
						MType: "gauge",
						Value: &testGauge,
					},
				},
			}
			result, err := storage.GetAll()
			if err != nil {
				t.Errorf("GetAll() failed: %v", err)
				return
			}
			assert.Equal(t, result, &tt.result)
		})
	}
}

func Benchmark_Set(b *testing.B) {

	mdb := NewMemStorage()
	metric := usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mdb.Set(&metric)
	}
}

func Benchmark_Get(b *testing.B) {

	mdb := NewMemStorage()
	metric := usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge}
	mdb.Set(&metric)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mdb.Get(&metric)
	}
}
