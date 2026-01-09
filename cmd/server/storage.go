package main

import (
	"errors"
	"fmt"
	"strconv"
)

type repositories interface {
	Update(kind, name, value string) error
}

func UpdateMetric(repo repositories, kind, name, value string) error {
	return repo.Update(kind, name, value)
}

type MemStorage struct {
	Value map[string]interface{}
}

func (m *MemStorage) Update(kind, name, value string) error {
	var err error

	if m.Value == nil {
		m.Value = make(map[string]interface{})
	}

	switch kind {
	case "gauge":
		m.Value[name], err = strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.New("Error gauge conversion")
		}
		fmt.Printf("gauge %s=%v\n", name, m.Value[name])
	case "counter":
		if _, ok := m.Value[name]; !ok {
			m.Value[name] = counter(0)
		}
		addition, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.New("Error counter conversion")
		}
		m.Value[name] = m.Value[name].(counter) + counter(addition)
		fmt.Printf("cntr %s=%v\n", name, m.Value[name])
	default:
		fmt.Printf("Unsupported value kind\n")
		return errors.New("Unsupported value kind")
	}
	return nil
}
