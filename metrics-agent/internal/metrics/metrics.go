package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
)

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

type metricsBatch struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMetricsBatch() *metricsBatch {
	return &metricsBatch{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func getRuntimeMetric(memstat *runtime.MemStats, name string) (float64, error) {
	v := reflect.ValueOf(*memstat)
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

func (m *metricsBatch) GetAllRuntimeMetrics(list []string) error {
	var r runtime.MemStats
	var err error
	runtime.ReadMemStats(&r)

	for _, s := range list {
		m.Gauge[s], err = getRuntimeMetric(&r, s)
		if err != nil {
			fmt.Printf("%s error: %v\n", s, err)
			return err
		} else {
			fmt.Printf("%s=%f\n", s, m.Gauge[s])
		}
	}
	return nil
}

func sendMetric(url, kind, name, value string) error {
	resp, err := http.Post(url+"/"+kind+"/"+name+"/"+value, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (m *metricsBatch) SendAllMetrics(url string) error {
	for k, v := range m.Gauge {
		if err := sendMetric(url, "gauge", k, strconv.FormatFloat(float64(v), 'f', -1, 64)); err != nil {
			return err
		}
	}
	for k, v := range m.Counter {
		if err := sendMetric(url, "counter", k, strconv.FormatInt(int64(v), 10)); err != nil {
			return err
		}
	}
	return nil
}
