package storage

import "strconv"

func GaugeToString(gauge float64) string {
	return strconv.FormatFloat(gauge, 'f', -1, 64)
}

func CounterToString(counter int64) string {
	return strconv.FormatInt(counter, 10)
}

func StringToGauge(gauge string) (float64, error) {
	result, err := strconv.ParseFloat(gauge, 64)
	if err != nil {
		return 0.0, err
	}
	return result, nil
}

func StringToCounter(counter string) (int64, error) {
	result, err := strconv.ParseInt(counter, 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}
