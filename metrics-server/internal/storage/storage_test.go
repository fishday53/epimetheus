package storage

import "testing"

func Test_GaugeToString(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		param  float64
		result string
	}{
		{
			name:   "GaugeToString",
			param:  0.0005,
			result: "0.0005",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GaugeToString(tt.param)
			if result != tt.result {
				t.Errorf("Wrong conversion. Wait for %s, got %s", tt.result, result)
			}
		})
	}
}

func Test_CounterToString(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		param  int64
		result string
	}{
		{
			name:   "CounterToString",
			param:  1000,
			result: "1000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CounterToString(tt.param)
			if result != tt.result {
				t.Errorf("Wrong conversion. Wait for %s, got %s", tt.result, result)
			}
		})
	}
}

func Test_StringToGauge(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		param  string
		result float64
	}{
		{
			name:   "GaugeToString",
			param:  "0.0005",
			result: 0.0005,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := StringToGauge(tt.param)
			if *result != tt.result {
				t.Errorf("Wrong conversion. Wait for %f, got %f", tt.result, *result)
			}
		})
	}
}

func Test_StringToCounter(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		param  string
		result int64
	}{
		{
			name:   "GaugeToString",
			param:  "1000",
			result: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := StringToCounter(tt.param)
			if *result != tt.result {
				t.Errorf("Wrong conversion. Wait for %d, got %d", tt.result, *result)
			}
		})
	}
}
