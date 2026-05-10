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
		config   string
		want     Config
	}{
		{
			testName: "Defaults",
			envs:     make(map[string]string),
			config:   "",
			want: Config{
				Addr:           "localhost:8080",
				ReportInterval: 10,
				PollInterval:   2,
				HashKey:        "",
				RateLimit:      1,
				CryptoKeyPath:  "",
				ConfigPath:     "",
				Transport:      "http",
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
				"CONFIG":          "",
				"TRANSPORT":       "grpc",
			},
			config: "",
			want: Config{
				Addr:           "example.com:1234",
				ReportInterval: 100,
				PollInterval:   50,
				HashKey:        "key",
				RateLimit:      10,
				CryptoKeyPath:  "/path/to/pub/key",
				ConfigPath:     "",
				Transport:      "grpc",
			},
		},
		{
			testName: "Config.json",
			envs: map[string]string{
				"CONFIG": os.TempDir() + "config-dyetegecvehyWenteekbonth6quadAys.json",
			},
			config: `
				{
					"address": "example.com:8888",
					"report_interval": 200,
					"poll_interval": 100,
					"hash_key": "key",
					"rate_limit": 10,
					"crypto_key": "/path/to/key.pem",
					"transport": "grpc"
				}
			`,
			want: Config{
				Addr:           "example.com:8888",
				ReportInterval: 200,
				PollInterval:   100,
				HashKey:        "key",
				RateLimit:      10,
				CryptoKeyPath:  "/path/to/key.pem",
				ConfigPath:     os.TempDir() + "config-dyetegecvehyWenteekbonth6quadAys.json",
				Transport:      "grpc",
			},
		},
		{
			testName: "Mixed",
			envs: map[string]string{
				"ADDRESS":       "example.com:1234",
				"POLL_INTERVAL": "50",
				"KEY":           "key",
				"CRYPTO_KEY":    "/path/to/pub/key",
				"CONFIG":        os.TempDir() + "config-dyetegecvehyWenteekbonth6quadAys.json",
				"TRANSPORT":     "http",
			},
			config: `
				{
					"address": "example.com:8888",
					"poll_interval": 100,
					"hash_key": "key",
					"rate_limit": 1000,
					"crypto_key": "/path/to/key.pem",
					"transport": "grpc"
				}
			`,
			want: Config{
				Addr:           "example.com:1234",
				ReportInterval: 10,
				PollInterval:   50,
				HashKey:        "key",
				RateLimit:      1000,
				CryptoKeyPath:  "/path/to/pub/key",
				ConfigPath:     os.TempDir() + "config-dyetegecvehyWenteekbonth6quadAys.json",
				Transport:      "http",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			if tt.config != "" {
				tmpFile, err := os.Create(tt.envs["CONFIG"])
				if err != nil {
					t.Fatalf("Generate config failed: %v", err)
				}
				defer tmpFile.Close()
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.config)
				if err != nil {
					t.Fatalf("Write config failed: %v", err)
				}
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
