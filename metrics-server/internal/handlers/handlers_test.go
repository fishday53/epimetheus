package handlers

import (
	"bytes"
	"encoding/json"
	"metrics-server/internal/storage"
	"metrics-server/internal/storage/memory"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var testCounter int64 = 527
var testGauge float64 = 0.00005

func Test_SetParam(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		metric storage.Metric
		want   want
	}{
		{
			name: "counter",
			metric: storage.Metric{
				ID:    "c1",
				MType: "counter",
				Delta: &testCounter,
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "gauge",
			metric: storage.Metric{
				ID:    "g1",
				MType: "gauge",
				Value: &testGauge,
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "bad kind",
			metric: storage.Metric{
				ID:    "g1",
				MType: "something",
				Value: &testGauge,
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "bad value",
			metric: storage.Metric{
				ID:    "g1",
				MType: "gauge",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no name",
			metric: storage.Metric{
				ID:    "",
				MType: "gauge",
				Value: &testGauge,
			},
			want: want{
				code: 404,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &AppContext{DB: memory.NewMemStorage()}
			r := chi.NewRouter()
			r.Post(`/update/`, ctx.SetParam)

			jsonData, _ := json.Marshal(tt.metric)
			reader := bytes.NewReader(jsonData)

			request := httptest.NewRequest(http.MethodPost, "/update/", reader)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()

			defer request.Body.Close()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_GetParam(t *testing.T) {
	type want struct {
		code   int
		answer string
	}
	tests := []struct {
		name    string
		storage memory.MemStorage
		request storage.Metric
		want    want
	}{
		{
			name: "Existent counter",
			request: storage.Metric{
				ID:    "c1",
				MType: "counter",
			},
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"c1": {MType: "counter", Delta: &testCounter}},
			},
			want: want{
				code:   200,
				answer: `{"id":"c1","type":"counter","delta":` + strconv.FormatInt(testCounter, 10) + `}`,
			},
		},
		{
			name: "Nonexistent counter",
			request: storage.Metric{
				ID:    "c2",
				MType: "counter",
			},
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"c1": {MType: "counter", Delta: &testCounter}},
			},
			want: want{
				code:   404,
				answer: "Value of c2 is absent\n",
			},
		},
		{
			name: "Existent gauge",
			request: storage.Metric{
				ID:    "g1",
				MType: "gauge",
			},
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"g1": {MType: "gauge", Value: &testGauge}},
			},
			want: want{
				code:   200,
				answer: `{"id":"g1","type":"gauge","value":` + strconv.FormatFloat(testGauge, 'f', -1, 64) + `}`,
			},
		},
		{
			name: "Nonexistent gauge",
			request: storage.Metric{
				ID:    "g2",
				MType: "gauge",
			},
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"g1": {MType: "gauge", Value: &testGauge}},
			},
			want: want{
				code:   404,
				answer: "Value of g2 is absent\n",
			},
		},
		{
			name: "Bad kind",
			request: storage.Metric{
				ID:    "g1",
				MType: "SomeWrongKind",
			},
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"g1": {MType: "gauge", Value: &testGauge}},
			},
			want: want{
				code:   404,
				answer: "Value of g1 is absent\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &AppContext{DB: &tt.storage}
			r := chi.NewRouter()
			r.Post(`/value/`, ctx.GetParam)

			jsonData, _ := json.Marshal(tt.request)
			reader := bytes.NewReader(jsonData)

			request := httptest.NewRequest(http.MethodPost, "/value/", reader)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()

			defer request.Body.Close()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.answer, w.Body.String())
		})
	}
}

func Test_getAllParams(t *testing.T) {
	type want struct {
		code   int
		answer string
	}
	tests := []struct {
		name    string
		storage memory.MemStorage
		request string
		want    want
	}{
		{
			name:    "Simple check",
			request: "/",
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{
					"g1": {MType: "gauge", Value: &testGauge},
					"c1": {MType: "counter", Delta: &testCounter},
				},
			},
			want: want{
				code:   200,
				answer: `[{"id":"g1","type":"gauge","value":` + strconv.FormatFloat(testGauge, 'f', -1, 64) + `},{"id":"c1","type":"counter","delta":` + strconv.FormatInt(testCounter, 10) + `}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &AppContext{DB: &tt.storage}
			r := chi.NewRouter()
			r.Get(`/`, ctx.GetAllParams)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()

			defer request.Body.Close()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.answer, w.Body.String())
		})
	}
}
