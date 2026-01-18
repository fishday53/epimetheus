package memory

import (
	"fmt"
	"log"
	"metrics-server/internal/storage"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *MemStorage) Set(kind, name, value string) error {
	var err error

	switch kind {
	case "gauge":
		m.Gauge[name], err = storage.StringToGauge(value)
		if err != nil {
			return fmt.Errorf("error gauge conversion: %v", err)
		}
		log.Printf("gauge %s=%v\n", name, m.Gauge[name])
	case "counter":
		if _, ok := m.Counter[name]; !ok {
			m.Counter[name] = 0
		}
		addition, err := storage.StringToCounter(value)
		if err != nil {
			return fmt.Errorf("error counter conversion: %v", err)
		}
		m.Counter[name] += addition
		log.Printf("cntr %s=%v\n", name, m.Counter[name])
	default:
		log.Printf("Unsupported value kind\n")
		return fmt.Errorf("unsupported value kind: %s", kind)
	}
	return nil
}

func (m *MemStorage) Get(kind, name string) (string, error) {
	switch kind {
	case "gauge":
		if result, ok := m.Gauge[name]; ok {
			return storage.GaugeToString(result), nil
		} else {
			return "", fmt.Errorf("gauge value for %s not found", name)
		}
	case "counter":
		if result, ok := m.Counter[name]; ok {
			return storage.CounterToString(result), nil
		} else {
			return "", fmt.Errorf("counter value for %s not found", name)
		}
	default:
		return "", fmt.Errorf("value %s has unsupported kind: %s", name, kind)
	}
}

func (m *MemStorage) GetAll() ([]string, error) {
	result := []string{}
	for k, v := range m.Gauge {
		result = append(result, k+":\t"+storage.GaugeToString(v)+"\n")
	}
	for k, v := range m.Counter {
		result = append(result, k+":\t"+storage.CounterToString(v)+"\n")
	}
	return result, nil
}
