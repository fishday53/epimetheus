// This part of storage package implements different metric kinds conversion to/from string.
package storage

import "strconv"

// GaugeToString converts gauge metric value to string.
func GaugeToString(gauge float64) string {
	return strconv.FormatFloat(gauge, 'f', -1, 64)
}

// CounterToString converts counter metric value to string.
func CounterToString(counter int64) string {
	return strconv.FormatInt(counter, 10)
}

// StringToGauge converts string metric value to gauge type.
func StringToGauge(gauge string) (*float64, error) {
	result, err := strconv.ParseFloat(gauge, 64)
	if err != nil {
		result = 0.0
	}
	return &result, err
}

// StringToCounter converts string metric value to counter type.
func StringToCounter(counter string) (*int64, error) {
	result, err := strconv.ParseInt(counter, 10, 64)
	if err != nil {
		result = 0
	}
	return &result, err
}
