package main

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
)

var metricList = []string{
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

type gauge float64
type counter int64

type MetricsBatch struct {
	Gauge   map[string]gauge
	Counter map[string]counter
}

func getRuntimeMetric(memstat *runtime.MemStats, name string) (gauge, error) {
	v := reflect.ValueOf(*memstat)
	fieldValue := v.FieldByName(name)
	if !fieldValue.IsValid() {
		return 0, errors.New("Value not found")
	}

	switch metric := fieldValue.Interface().(type) {
	case float64:
		return gauge(metric), nil
	case uint32:
		return gauge(metric), nil
	case uint64:
		return gauge(metric), nil
	default:
		// return fieldValue.Interface().(float64), nil
		return 0, fmt.Errorf("Unknown type %v", metric)
	}
}

func (m *MetricsBatch) getAllRuntimeMetrics(list []string) error {
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
	_, err := http.Post(url+"/"+kind+"/"+name+"/"+value, "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

func (m *MetricsBatch) sendAllMetrics(url string) error {
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
