package memory

import (
	"fmt"
	"log"
	"metrics-server/internal/storage"
)

type MetricParam struct {
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MemStorage struct {
	Name    string
	Metrics map[string]MetricParam
}

func NewMemStorage(name string) *MemStorage {
	return &MemStorage{
		Name:    name,
		Metrics: make(map[string]MetricParam),
	}
}

func (m *MemStorage) Set(metric *storage.Metric) (*storage.Metric, error) {

	result := storage.Metric{
		ID:    metric.ID,
		MType: metric.MType,
	}

	switch metric.MType {

	case "gauge":
		if _, ok := m.Metrics[metric.ID]; ok {
			if m.Metrics[metric.ID].MType != "gauge" {
				log.Printf("Value type changing is not enabled\n")
				return nil, fmt.Errorf("value type changing is not enabled: %s", metric.MType)
			}
		}
		if metric.Value == nil {
			log.Printf("Value is nil\n")
			return nil, fmt.Errorf("value is nil")
		}

		m.Metrics[metric.ID] = MetricParam{MType: "gauge", Value: metric.Value}
		result.Value = m.Metrics[metric.ID].Value
		log.Printf("gauge %s=%v\n", metric.ID, m.Metrics[metric.ID].Value)

	case "counter":
		if _, ok := m.Metrics[metric.ID]; !ok {
			var initialDelta int64 = 0
			m.Metrics[metric.ID] = MetricParam{MType: "counter", Delta: &initialDelta}
		} else {
			if m.Metrics[metric.ID].MType != "counter" {
				log.Printf("Value type changing is not enabled\n")
				return nil, fmt.Errorf("value type changing is not enabled: %s", metric.MType)
			}
		}
		if metric.Delta == nil {
			log.Printf("Delta is nil\n")
			return nil, fmt.Errorf("delta is nil")
		}

		//incr := *metric.Delta
		//*m.Metrics[metric.ID].Delta += incr
		*m.Metrics[metric.ID].Delta += *metric.Delta
		//m.Metrics[metric.ID] = MetricParam{MType: "counter", Delta: metric.Delta}
		result.Delta = m.Metrics[metric.ID].Delta
		log.Printf("storage: %v, cntr %s=%v\n", m, metric.ID, *m.Metrics[metric.ID].Delta)

	default:
		log.Printf("Unsupported value kind\n")
		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	}

	return &result, nil
}

func (m *MemStorage) Get(metric *storage.Metric) (*storage.Metric, error) {

	if _, ok := m.Metrics[metric.ID]; !ok {
		return nil, fmt.Errorf("%s not found", metric.ID)
	}
	if m.Metrics[metric.ID].MType != metric.MType {
		log.Printf("Value type is wrong\n")
		return nil, fmt.Errorf("value type is wrong: %s", metric.MType)
	}

	switch metric.MType {
	case "gauge":
		metric.Value = m.Metrics[metric.ID].Value
	case "counter":
		metric.Delta = m.Metrics[metric.ID].Delta
	default:
		return nil, fmt.Errorf("value %s has unsupported kind: %s", metric.ID, metric.MType)
	}

	return metric, nil
}

func (m *MemStorage) GetAll() (*[]storage.Metric, error) {
	result := []storage.Metric{}
	for k, v := range m.Metrics {
		result = append(result, storage.Metric{
			ID:    k,
			MType: v.MType,
			Delta: v.Delta,
			Value: v.Value,
		})
	}
	return &result, nil
}
