package postgres

import (
	"database/sql"
	"fmt"
	"metrics-server/internal/usecase"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const table = "metrics"

type PsqlStorage struct {
	DB *sql.DB
}

func NewPsqlStorage(dsn string) (*PsqlStorage, error) {
	p := PsqlStorage{}
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
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		mtype VARCHAR(255) NOT NULL,
		delta BIGINT DEFAULT 0,
		value FLOAT8 DEFAULT 0.0)
	`, table)
	_, err := p.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("cannot create table %s: %v", query, err)
	}
	return nil
}

func (p *PsqlStorage) Set(metric *usecase.Metric) (*usecase.Metric, error) {

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
		_, err := p.DB.Exec(query, result.ID, result.MType, *result.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot set value: %v", err)
		}
		result.Value = metric.Value

	case "counter":

		*result.Delta += *metric.Delta

		query := fmt.Sprintf(`
		INSERT INTO %s (id, mtype, delta) VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET delta = $3`, table)
		_, err := p.DB.Exec(query, result.ID, result.MType, *result.Delta)
		if err != nil {
			return nil, fmt.Errorf("cannot set delta: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	}

	return result, nil
}

func (p *PsqlStorage) Get(metric *usecase.Metric) (*usecase.Metric, error) {
	var delta int64
	var value float64

	query := fmt.Sprintf("SELECT delta, value FROM %s WHERE id = $1 AND mtype = $2", table)

	row := p.DB.QueryRow(query, metric.ID, metric.MType)
	err := row.Scan(&delta, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s not found", metric.ID)
		}
		return nil, fmt.Errorf("sql query error: %v", err)
	}

	metric.Delta = &delta
	metric.Value = &value

	return metric, nil
}

func (p *PsqlStorage) GetAll() (*[]usecase.Metric, error) {
	result := []usecase.Metric{}

	query := fmt.Sprintf("SELECT id, mtype, delta, value FROM %s", table)

	rows, err := p.DB.Query(query)
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
