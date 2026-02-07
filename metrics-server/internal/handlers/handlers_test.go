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
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var testCounter int64 = 527
var testGauge float64 = 0.00005

func Test_SetParam(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "counter",
			request: "/update/counter/c1/527",
			want: want{
				code: 200,
			},
		},
		{
			name:    "gauge",
			request: "/update/gauge/g1/-0.1",
			want: want{
				code: 200,
			},
		},
		{
			name:    "bad mtype",
			request: "/update/something/g1/-0.1",
			want: want{
				code: 400,
			},
		},
		{
			name:    "bad value",
			request: "/update/gauge/g2/b",
			want: want{
				code: 400,
			},
		},
		{
			name:    "no name",
			request: "/update/gauge/b",
			want: want{
				code: 404,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &AppContext{DB: memory.NewMemStorage()}
			r := chi.NewRouter()
			r.Post(`/update/{mtype}/{name}/{value}`, ctx.SetParam)

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
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
		request string
		want    want
	}{
		{
			name:    "Existent counter",
			request: "/value/counter/c1",
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"c1": {MType: "counter", Delta: &testCounter}},
			},
			want: want{
				code:   200,
				answer: storage.CounterToString(testCounter) + "\n",
			},
		},
		{
			name:    "Nonexistent counter",
			request: "/value/counter/c2",
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"c1": {MType: "counter", Delta: &testCounter}},
			},
			want: want{
				code:   404,
				answer: "Value of c2 is absent\n",
			},
		},
		{
			name:    "Existent gauge",
			request: "/value/gauge/g1",
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"g1": {MType: "gauge", Value: &testGauge}},
			},
			want: want{
				code:   200,
				answer: storage.GaugeToString(testGauge) + "\n",
			},
		},
		{
			name:    "Nonexistent gauge",
			request: "/value/gauge/g2",
			storage: memory.MemStorage{
				Metrics: map[string]memory.MetricParam{"g1": {MType: "gauge", Value: &testGauge}},
			},
			want: want{
				code:   404,
				answer: "Value of g2 is absent\n",
			},
		},
		{
			name:    "Bad mtype",
			request: "/value/SomeWrongMType/g1",
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
			r.Get(`/value/{mtype}/{name}`, ctx.GetParam)

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
				answer: "g1:\t" + storage.GaugeToString(testGauge) + "\nc1:\t" + storage.CounterToString(testCounter) + "\n",
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

func Test_SetParamJSON(t *testing.T) {
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
			name: "bad mtype",
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
			r.Post(`/update/`, ctx.SetParamJSON)

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

func Test_GetParamJSON(t *testing.T) {
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
			name: "Bad mtype",
			request: storage.Metric{
				ID:    "g1",
				MType: "SomeWrongNType",
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
			r.Post(`/value/`, ctx.GetParamJSON)

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

func Test_getAllParamsJSON(t *testing.T) {
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
			r.Get(`/`, ctx.GetAllParamsJSON)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()

			defer request.Body.Close()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			var data1, data2 interface{}
			json.Unmarshal([]byte(tt.want.answer), &data1)
			json.Unmarshal(w.Body.Bytes(), &data2)
			diff := cmp.Diff(data1, data2)
			assert.Equal(t, diff, "")
		})
	}
}
