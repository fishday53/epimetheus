package storage

//var InitialDelta int64 = 0

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type Repositories interface {
	Set(metric *Metric) (*Metric, error)
	Get(metric *Metric) (*Metric, error)
	GetAll() (*[]Metric, error)
}
