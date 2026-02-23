package postgres

import (
	"database/sql"
	"fmt"
	"metrics-server/internal/storage"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const table = "metrics"

type PsqlStorage struct {
	Name string
	DB   *sql.DB
}

func NewPsqlStorage(name string, dsn string) (*PsqlStorage, error) {
	p := PsqlStorage{Name: name}
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
	query := `
	CREATE TABLE IF NOT EXISTS $1 (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		mtype VARCHAR(255) NOT NULL,
		delta BIGINT DEFAULT 0,
		value FLOAT8 DEFAULT 0.0
	);`
	_, err := p.DB.Exec(query, table)
	if err != nil {
		return fmt.Errorf("cannot create table: %v", err)
	}
	return nil
}

func (p *PsqlStorage) Set(metric *storage.Metric) (*storage.Metric, error) {

	result, err := p.Get(metric)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			result = &storage.Metric{
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

		*result.Value = *metric.Value

		query := `
		INSERT INTO $1 (id, value) VALUES ($2, $3)
		ON CONFLICT (id)
		DO UPDATE SET value = $3 WHERE id = $2;`
		_, err := p.DB.Exec(query, table, result.ID, *result.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot set value: %v", err)
		}
		result.Value = metric.Value

	case "counter":

		*result.Delta += *metric.Delta

		query := `
		INSERT INTO $1 (id, delta) VALUES ($2, $3)
		ON CONFLICT (id)
		DO UPDATE SET delta = $3 WHERE id = $2;`
		_, err := p.DB.Exec(query, table, result.ID, *result.Delta)
		if err != nil {
			return nil, fmt.Errorf("cannot set delta: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	}

	return result, nil
}

func (p *PsqlStorage) Get(metric *storage.Metric) (*storage.Metric, error) {

	query := "SELECT delta, value FROM $1 WHERE id = $2 and mtype = $3"

	row := p.DB.QueryRow(query, table, metric.ID, metric.MType)
	err := row.Scan(metric.Delta, metric.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s not found", metric.ID)
		}
		return nil, fmt.Errorf("sql query error: %v", err)
	}

	return metric, nil
}

func (p *PsqlStorage) GetAll() (*[]storage.Metric, error) {
	result := []storage.Metric{}

	query := "SELECT id, mtype, delta, value FROM $1"

	rows, err := p.DB.Query(query, table)
	if err != nil {
		return nil, fmt.Errorf("error in query for all metrics: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		metric := storage.Metric{}
		if err := rows.Scan(&metric.ID, &metric.MType, metric.Delta, metric.Value); err != nil {
			return nil, fmt.Errorf("cannot process a row: %v", err)
		}
		result = append(result, metric)
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
