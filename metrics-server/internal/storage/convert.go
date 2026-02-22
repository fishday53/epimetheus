package storage

import "strconv"

func GaugeToString(gauge float64) string {
	return strconv.FormatFloat(gauge, 'f', -1, 64)
}

func CounterToString(counter int64) string {
	return strconv.FormatInt(counter, 10)
}

func StringToGauge(gauge string) (*float64, error) {
	result, err := strconv.ParseFloat(gauge, 64)
	if err != nil {
		result = 0.0
	}
	return &result, err
}

func StringToCounter(counter string) (*int64, error) {
	result, err := strconv.ParseInt(counter, 10, 64)
	if err != nil {
		result = 0
	}
	return &result, err
}
