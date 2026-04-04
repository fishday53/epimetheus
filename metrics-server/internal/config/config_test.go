package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {
	tests := []struct {
		testName string
		envs     map[string]string
		want     Config
	}{
		{
			testName: "Defaults",
			envs:     make(map[string]string),
			want: Config{
				Addr:            "localhost:8080",
				StoreInterval:   300,
				FileStoragePath: "metrics.dmp",
				Restore:         true,
				DSN:             "",
				HashKey:         "",
			},
		},
		{
			testName: "Envs",
			envs: map[string]string{
				"ADDRESS":           "example.com:1234",
				"STORE_INTERVAL":    "100",
				"FILE_STORAGE_PATH": "/tmp/metrics.dmp",
				"RESTORE":           "false",
				"DATABASE_DSN":      "user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable",
				"KEY":               "secret",
			},
			want: Config{
				Addr:            "example.com:1234",
				StoreInterval:   100,
				FileStoragePath: "/tmp/metrics.dmp",
				Restore:         false,
				DSN:             "user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable",
				HashKey:         "secret",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			var cfg Config
			err := cfg.Get()
			if err != nil {
				t.Fatalf("Method Get() failed: %v", err)
			}
			assert.Equal(t, tt.want, cfg)
		})
	}
}
