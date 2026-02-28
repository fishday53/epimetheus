package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"metrics-server/internal/usecase"
	"os"
)

type MetricParam struct {
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MemStorage struct {
	Metrics map[string]MetricParam
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]MetricParam),
	}
}

func (m *MemStorage) Set(metric *usecase.Metric) (*usecase.Metric, error) {

	result := usecase.Metric{
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

		*m.Metrics[metric.ID].Delta += *metric.Delta
		result.Delta = m.Metrics[metric.ID].Delta

	default:
		log.Printf("Unsupported value kind\n")
		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	}

	return &result, nil
}

func (m *MemStorage) Get(metric *usecase.Metric) (*usecase.Metric, error) {

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

func (m *MemStorage) GetAll() (*[]usecase.Metric, error) {
	result := []usecase.Metric{}
	for k, v := range m.Metrics {
		result = append(result, usecase.Metric{
			ID:    k,
			MType: v.MType,
			Delta: v.Delta,
			Value: v.Value,
		})
	}
	return &result, nil
}

func (m *MemStorage) Dump(filepath string) error {
	data, err := json.MarshalIndent(m.Metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("error in dump marshaller: %v", err)
	}
	return os.WriteFile(filepath, data, 0666)
}

func (m *MemStorage) Restore(filepath string) error {
	var data []byte
	var err error
	data, err = os.ReadFile(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("File %s does not exist\n", filepath)
			return nil
		}
		return fmt.Errorf("cannot read file %s for restoration: %v", filepath, err)
	}
	if err = json.Unmarshal(data, &m.Metrics); err != nil {
		return fmt.Errorf("cannot unmarshal file %s data for restoration: %v", filepath, err)
	}
	return nil
}
