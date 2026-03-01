package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"metrics-server/internal/usecase"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const table = "metrics"

type PsqlStorage struct {
	DB              *sql.DB
	BackOffSchedule *[]time.Duration
}

var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func NewPsqlStorage(dsn string) (*PsqlStorage, error) {
	p := PsqlStorage{BackOffSchedule: &backoffSchedule}
	var err error

	p.DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot create pgx: %v", err)
	}

	if err := p.Migrate(); err != nil {
		return nil, fmt.Errorf("cannot migrate db: %v", err)
	}

	return &p, nil
}

func (p *PsqlStorage) Migrate() error {
	var err error
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		mtype VARCHAR(255) NOT NULL,
		delta BIGINT DEFAULT 0,
		value FLOAT8 DEFAULT 0.0)
	`, table)

	for _, backoff := range *p.BackOffSchedule {
		if _, err = p.DB.Exec(query); err == nil {
			return nil
		}
		time.Sleep(backoff)
	}

	return fmt.Errorf("cannot create table %s: %v", query, err)
}

func (p *PsqlStorage) Set(metric *usecase.Metric) (*usecase.Metric, error) {
	var err error

	result, err := p.Get(metric)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			result = &usecase.Metric{
				ID:    metric.ID,
				MType: metric.MType,
			}
			if metric.MType == "counter" {
				var initialDelta int64 = 0
				result.Delta = &initialDelta
			}
		} else {
			return nil, fmt.Errorf("cannot check metric: %v", err)
		}
	}

	if metric.MType != result.MType {
		return nil, fmt.Errorf("value type changing is not enabled: %s", metric.MType)
	}

	switch metric.MType {

	case "gauge":

		result.Value = metric.Value

		query := fmt.Sprintf(`
		INSERT INTO %s (id, mtype, value) VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET value = $3`, table)
		for _, backoff := range *p.BackOffSchedule {
			if _, err := p.DB.Exec(query, result.ID, result.MType, *result.Value); err == nil {
				return result, nil
			}
			time.Sleep(backoff)
		}
		return nil, fmt.Errorf("cannot set value: %v", err)

	case "counter":

		*result.Delta += *metric.Delta

		query := fmt.Sprintf(`
		INSERT INTO %s (id, mtype, delta) VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET delta = $3`, table)
		for _, backoff := range *p.BackOffSchedule {
			if _, err := p.DB.Exec(query, result.ID, result.MType, *result.Delta); err == nil {
				return result, nil
			}
			time.Sleep(backoff)
		}
		return nil, fmt.Errorf("cannot set delta: %v", err)

	default:
		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	}
}

func (p *PsqlStorage) Get(metric *usecase.Metric) (*usecase.Metric, error) {
	var err error
	var delta int64
	var value float64
	result := usecase.Metric{
		ID:    metric.ID,
		MType: metric.MType,
		Delta: &delta,
		Value: &value,
	}

	query := fmt.Sprintf("SELECT delta, value FROM %s WHERE id = $1 AND mtype = $2", table)

	for _, backoff := range *p.BackOffSchedule {
		row := p.DB.QueryRow(query, metric.ID, metric.MType)
		err = row.Scan(result.Delta, result.Value)
		if err == nil {
			return &result, nil
		} else {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("%s not found", metric.ID)
			}
		}
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("sql query error: %v", err)
}

func (p *PsqlStorage) GetAll() (*[]usecase.Metric, error) {
	var err error
	var rows *sql.Rows
	result := []usecase.Metric{}

	query := fmt.Sprintf("SELECT id, mtype, delta, value FROM %s", table)

	for _, backoff := range *p.BackOffSchedule {
		if rows, err = p.DB.Query(query); err != nil {
			time.Sleep(backoff)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error in query for all metrics: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var mtype string
		var delta int64
		var value float64
		if err := rows.Scan(&id, &mtype, &delta, &value); err != nil {
			return nil, fmt.Errorf("cannot process a row: %v", err)
		}

		result = append(result, usecase.Metric{ID: id, MType: mtype, Delta: &delta, Value: &value})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot process all rows: %v", err)
	}

	return &result, nil
}

func (p *PsqlStorage) Dump(filepath string) error {
	// not implemented
	return nil
}

func (p *PsqlStorage) Restore(filepath string) error {
	// not implemented
	return nil
}

func (p *PsqlStorage) Ping() error {
	var err error
	for _, backoff := range *p.BackOffSchedule {
		c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = p.DB.PingContext(c); err == nil {
			return nil
		}
		time.Sleep(backoff)
	}
	return fmt.Errorf("Cannot ping DB: %v", err)
}
