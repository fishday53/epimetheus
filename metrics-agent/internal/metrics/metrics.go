package metrics

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var MetricList = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

func GetRuntimeMetric(name string) (float64, error) {
	var r runtime.MemStats
	runtime.ReadMemStats(&r)

	v := reflect.ValueOf(r)
	fieldValue := v.FieldByName(name)
	if !fieldValue.IsValid() {
		return 0, errors.New("value not found")
	}

	switch metric := fieldValue.Interface().(type) {
	case float64:
		return metric, nil
	case uint32:
		return float64(metric), nil
	case uint64:
		return float64(metric), nil
	default:
		return 0, fmt.Errorf("unknown type %v", metric)
	}
}
