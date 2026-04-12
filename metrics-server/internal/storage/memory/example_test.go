package memory

import (
	"fmt"
	"metrics-server/internal/usecase"
)

func ExampleMemStorage_Set() {
	storage := NewMemStorage()
	testGauge := 0.1
	metric := usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge}

	result, _ := storage.Set(&metric)

	fmt.Println(result.ID, *result.Value)

	// Output:
	// g1 0.1
}

func ExampleMemStorage_Get() {
	storage := NewMemStorage()
	testGauge := 0.1

	storage.Metrics["g1"] = MetricParam{MType: "gauge", Value: &testGauge}

	metric := usecase.Metric{ID: "g1", MType: "gauge"}

	result, _ := storage.Get(&metric)

	fmt.Println(result.ID, *result.Value)

	// Output:
	// g1 0.1
}
