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
				Addr:           "localhost:8080",
				ReportInterval: 10,
				PollInterval:   2,
				HashKey:        "",
				RateLimit:      1,
				CryptoKeyPath:  "",
			},
		},
		{
			testName: "Envs",
			envs: map[string]string{
				"ADDRESS":         "example.com:1234",
				"REPORT_INTERVAL": "100",
				"POLL_INTERVAL":   "50",
				"KEY":             "key",
				"RATE_LIMIT":      "10",
				"CRYPTO_KEY":      "/path/to/pub/key",
			},
			want: Config{
				Addr:           "example.com:1234",
				ReportInterval: 100,
				PollInterval:   50,
				HashKey:        "key",
				RateLimit:      10,
				CryptoKeyPath:  "/path/to/pub/key",
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
