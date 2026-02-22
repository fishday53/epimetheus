package postgres

import (
	"database/sql"
	"fmt"
	"metrics-server/internal/storage"

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

	result := storage.Metric{
		ID:    metric.ID,
		MType: metric.MType,
	}

	// not implemented yet
	// 	switch metric.MType {

	// 	case "gauge":

	// 		query := `
	// INSERT INTO ? (id, mtype, delta) VALUES ($1, $2, $3)
	// ON CONFLICT (id)
	// DO UPDATE SET delta = $3;`

	// 		if _, ok := m.Metrics[metric.ID]; ok {
	// 			if m.Metrics[metric.ID].MType != "gauge" {
	// 				log.Printf("Value type changing is not enabled\n")
	// 				return nil, fmt.Errorf("value type changing is not enabled: %s", metric.MType)
	// 			}
	// 		}
	// 		if metric.Value == nil {
	// 			log.Printf("Value is nil\n")
	// 			return nil, fmt.Errorf("value is nil")
	// 		}

	// 		m.Metrics[metric.ID] = MetricParam{MType: "gauge", Value: metric.Value}
	// 		result.Value = m.Metrics[metric.ID].Value

	// 	case "counter":
	// 		if _, ok := m.Metrics[metric.ID]; !ok {
	// 			var initialDelta int64 = 0
	// 			m.Metrics[metric.ID] = MetricParam{MType: "counter", Delta: &initialDelta}
	// 		} else {
	// 			if m.Metrics[metric.ID].MType != "counter" {
	// 				log.Printf("Value type changing is not enabled\n")
	// 				return nil, fmt.Errorf("value type changing is not enabled: %s", metric.MType)
	// 			}
	// 		}
	// 		if metric.Delta == nil {
	// 			log.Printf("Delta is nil\n")
	// 			return nil, fmt.Errorf("delta is nil")
	// 		}

	// 		*m.Metrics[metric.ID].Delta += *metric.Delta
	// 		result.Delta = m.Metrics[metric.ID].Delta

	// 	default:
	// 		log.Printf("Unsupported value kind\n")
	// 		return nil, fmt.Errorf("unsupported value kind: %s", metric.MType)
	// 	}

	return &result, nil
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
