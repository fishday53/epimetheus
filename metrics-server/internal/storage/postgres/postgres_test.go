package postgres

import (
	"database/sql"
	"testing"

	"metrics-server/internal/usecase"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testCounter   int64   = 527
	testGauge     float64 = 0.00005
	resultCounter int64   = testCounter + testCounter
	resultGauge   float64 = testGauge
)

func Test_Set(t *testing.T) {
	tests := []struct {
		name       string
		metric     usecase.Metric
		result     usecase.Metric
		wantErr    bool
		setupMocks func(mock sqlmock.Sqlmock, metric usecase.Metric)
	}{
		{
			name:    "Set new counter",
			metric:  usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter},
			result:  usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter},
			wantErr: false,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				// No rows for counter
				mock.ExpectQuery("SELECT delta FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnError(sql.ErrNoRows)

				// Insert one
				mock.ExpectExec("INSERT INTO "+table).
					WithArgs(metric.ID, metric.MType, testCounter).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:    "Update counter",
			metric:  usecase.Metric{ID: "c2", MType: "counter", Delta: &testCounter},
			result:  usecase.Metric{ID: "c2", MType: "counter", Delta: &resultCounter},
			wantErr: false,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				// Counter already exists
				rows := sqlmock.NewRows([]string{"delta"}).AddRow(testCounter)
				mock.ExpectQuery("SELECT delta FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnRows(rows)

				// Insert one
				mock.ExpectExec("INSERT INTO metrics").
					WithArgs(metric.ID, metric.MType, resultCounter).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:    "Set gauge",
			metric:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			result:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			wantErr: false,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				// No rows for gauge
				mock.ExpectQuery("SELECT value FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnError(sql.ErrNoRows)

				// Insert one
				mock.ExpectExec("INSERT INTO "+table).
					WithArgs(metric.ID, metric.MType, testGauge).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:    "Change type for counter",
			metric:  usecase.Metric{ID: "c1", MType: "gauge", Value: &testGauge},
			wantErr: true,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				// No rows for gauge
				mock.ExpectQuery("SELECT value FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnError(sql.ErrNoRows)

				// Counter already exists
				rows := sqlmock.NewRows([]string{"delta"}).AddRow(testCounter)
				mock.ExpectQuery("SELECT delta FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, "counter").
					WillReturnRows(rows)
			},
		},
		{
			name:    "Change type for gauge",
			metric:  usecase.Metric{ID: "g1", MType: "counter", Delta: &testCounter},
			wantErr: true,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				// No rows for counter
				mock.ExpectQuery("SELECT delta FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnError(sql.ErrNoRows)

				// Gauge already exists
				rows := sqlmock.NewRows([]string{"value"}).AddRow(testGauge)
				mock.ExpectQuery("SELECT value FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, "gauge").
					WillReturnRows(rows)
			},
		},
		{
			name:       "Unsupported type",
			metric:     usecase.Metric{ID: "x1", MType: "unsupported", Delta: &testCounter},
			wantErr:    true,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMocks(mock, tt.metric)

			storage := &PsqlStorage{
				DB: db,
			}

			result, err := storage.Set(&tt.metric)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("Set() succeeded unexpectedly")
				return
			}
			assert.Equal(t, *result, tt.result)
		})
	}
}

func Test_Get(t *testing.T) {
	tests := []struct {
		name       string
		metric     usecase.Metric
		result     usecase.Metric
		wantErr    bool
		setupMocks func(mock sqlmock.Sqlmock, metric usecase.Metric)
	}{
		{
			name:    "Get counter",
			metric:  usecase.Metric{ID: "c1", MType: "counter"},
			result:  usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter},
			wantErr: false,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				rows := sqlmock.NewRows([]string{"delta"}).AddRow(testCounter)
				mock.ExpectQuery("SELECT delta FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnRows(rows)
			},
		},
		{
			name:    "Get gauge",
			metric:  usecase.Metric{ID: "g1", MType: "gauge"},
			result:  usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge},
			wantErr: false,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				rows := sqlmock.NewRows([]string{"value"}).AddRow(testGauge)
				mock.ExpectQuery("SELECT value FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnRows(rows)
			},
		},
		{
			name:    "Not found",
			metric:  usecase.Metric{ID: "g1", MType: "gauge"},
			wantErr: true,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {
				mock.ExpectQuery("SELECT value FROM "+table+" WHERE id = \\$1 AND mtype = \\$2").
					WithArgs(metric.ID, metric.MType).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:       "Unsupported type",
			metric:     usecase.Metric{ID: "x1", MType: "unsupported", Delta: &testCounter},
			wantErr:    true,
			setupMocks: func(mock sqlmock.Sqlmock, metric usecase.Metric) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMocks(mock, tt.metric)

			storage := &PsqlStorage{
				DB: db,
			}

			result, err := storage.Get(&tt.metric)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("Set() succeeded unexpectedly")
				return
			}
			assert.Equal(t, *result, tt.result)
		})
	}
}

func Test_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &PsqlStorage{
		DB: db,
	}

	var (
		zeroCounter int64
		zeroGauge   float64
	)

	c1 := usecase.Metric{ID: "c1", MType: "counter", Delta: &testCounter, Value: &zeroGauge}
	g1 := usecase.Metric{ID: "g1", MType: "gauge", Value: &testGauge, Delta: &zeroCounter}
	rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
		AddRow("g1", "gauge", 0.0, testGauge).
		AddRow("c1", "counter", testCounter, 0)

	mock.ExpectQuery("SELECT id, mtype, delta, value FROM " + table).
		WillReturnRows(rows)

	result, err := storage.GetAll()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)

	metrics := *result
	assert.Equal(t, g1, metrics[0])
	assert.Equal(t, c1, metrics[1])
}
