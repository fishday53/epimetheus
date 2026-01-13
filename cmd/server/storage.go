package main

import (
	"fmt"
	"strconv"
)

type repositories interface {
	Set(kind, name, value string) error
	Get(kind, name string) (string, error)
	GetAll() ([]string, error)
}

func setMetric(repo repositories, kind, name, value string) error {
	return repo.Set(kind, name, value)
}

func getMetric(repo repositories, kind, name string) (string, error) {
	return repo.Get(kind, name)
}

func getAllMetrics(repo repositories) ([]string, error) {
	return repo.GetAll()
}

type memStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *memStorage {
	return &memStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func gaugeToString(gauge float64) string {
	return strconv.FormatFloat(gauge, 'f', -1, 64)
}

func counterToString(counter int64) string {
	return strconv.FormatInt(counter, 10)
}

func stringToGauge(gauge string) (float64, error) {
	result, err := strconv.ParseFloat(gauge, 64)
	if err != nil {
		return 0.0, err
	}
	return result, nil
}

func stringToCounter(counter string) (int64, error) {
	result, err := strconv.ParseInt(counter, 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (m *memStorage) Set(kind, name, value string) error {
	var err error

	switch kind {
	case "gauge":
		m.Gauge[name], err = stringToGauge(value)
		if err != nil {
			return fmt.Errorf("error gauge conversion: %v", err)
		}
		fmt.Printf("gauge %s=%v\n", name, m.Gauge[name])
	case "counter":
		if _, ok := m.Counter[name]; !ok {
			m.Counter[name] = 0
		}
		addition, err := stringToCounter(value)
		if err != nil {
			return fmt.Errorf("error counter conversion: %v", err)
		}
		m.Counter[name] += addition
		fmt.Printf("cntr %s=%v\n", name, m.Counter[name])
	default:
		fmt.Printf("Unsupported value kind\n")
		return fmt.Errorf("unsupported value kind: %s", kind)
	}
	return nil
}

func (m *memStorage) Get(kind, name string) (string, error) {
	switch kind {
	case "gauge":
		if result, ok := m.Gauge[name]; ok {
			return gaugeToString(result), nil
		} else {
			return "", fmt.Errorf("gauge value for %s not found", name)
		}
	case "counter":
		if result, ok := m.Counter[name]; ok {
			return counterToString(result), nil
		} else {
			return "", fmt.Errorf("counter value for %s not found", name)
		}
	default:
		return "", fmt.Errorf("value %s has unsupported kind: %s", name, kind)
	}
}

func (m *memStorage) GetAll() ([]string, error) {
	result := []string{}
	for k, v := range m.Gauge {
		result = append(result, k+":\t"+gaugeToString(v)+"\n")
	}
	for k, v := range m.Counter {
		result = append(result, k+":\t"+counterToString(v)+"\n")
	}
	return result, nil
}
