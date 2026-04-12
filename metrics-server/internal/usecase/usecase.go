// Package usecase contains main business logic.
package usecase

type (
	// Metric is the metric structure.
	Metric struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}

	// Repository is an universal inreface for different metric storages.
	Repository interface {
		Set(metric *Metric) (*Metric, error)
		Get(metric *Metric) (*Metric, error)
		GetAll() (*[]Metric, error)
		Dump(filepath string) error
		Restore(filepath string) error
		Ping() error
	}
)
