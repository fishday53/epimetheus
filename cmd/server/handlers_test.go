package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func Test_setParam(t *testing.T) {
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
			name:    "bad kind",
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
			Storage = NewMemStorage()
			r := chi.NewRouter()
			r.Post(`/update/{kind}/{name}/{value}`, setParam)

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_getParam(t *testing.T) {
	type want struct {
		code   int
		answer string
	}
	tests := []struct {
		name    string
		storage memStorage
		request string
		want    want
	}{
		{
			name:    "Existent counter",
			request: "/value/counter/c1",
			storage: memStorage{
				Counter: map[string]int64{"c1": 527},
			},
			want: want{
				code:   200,
				answer: "c1:\t527\n",
			},
		},
		{
			name:    "Nonexistent counter",
			request: "/value/counter/c2",
			storage: memStorage{
				Counter: map[string]int64{"c1": 527},
			},
			want: want{
				code:   404,
				answer: "Value of c2 is absent\n",
			},
		},
		{
			name:    "Existent gauge",
			request: "/value/gauge/g1",
			storage: memStorage{
				Gauge: map[string]float64{"g1": 0.00005},
			},
			want: want{
				code:   200,
				answer: "g1:\t0.00005\n",
			},
		},
		{
			name:    "Nonexistent gauge",
			request: "/value/gauge/g2",
			storage: memStorage{
				Gauge: map[string]float64{"g1": 0.00005},
			},
			want: want{
				code:   404,
				answer: "Value of g2 is absent\n",
			},
		},
		{
			name:    "Bad kind",
			request: "/value/SomeWrongKind/g1",
			storage: memStorage{
				Gauge: map[string]float64{"g1": 0.00005},
			},
			want: want{
				code:   404,
				answer: "Value of g1 is absent\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Storage = &tt.storage
			r := chi.NewRouter()
			r.Get(`/value/{kind}/{name}`, getParam)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
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
		storage memStorage
		request string
		want    want
	}{
		{
			name:    "Simple check",
			request: "/",
			storage: memStorage{
				Counter: map[string]int64{"c1": 527},
				Gauge:   map[string]float64{"g1": 0.00005},
			},
			want: want{
				code:   200,
				answer: "g1:\t0.00005\nc1:\t527\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Storage = &tt.storage
			r := chi.NewRouter()
			r.Get(`/`, getAllParams)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.answer, w.Body.String())
		})
	}
}
