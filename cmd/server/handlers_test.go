package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			separator := "/"
			substrings := strings.Split(tt.request, separator)
			if len(substrings) > 4 {
				request.SetPathValue("kind", substrings[2])
				request.SetPathValue("name", substrings[3])
				request.SetPathValue("value", substrings[4])
			}
			// for i, s := range substrings {
			// 	fmt.Printf("%d)%s\n", i, s)
			// }

			setParam(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
