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
				Addr:            "localhost:8080",
				StoreInterval:   300,
				FileStoragePath: "metrics.dmp",
				Restore:         true,
				DSN:             "",
				HashKey:         "",
				CryptoKeyPath:   "",
				TrustedSubnet:   "",
				ConfigPath:      "",
				Transport:       "http",
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
				"CRYPTO_KEY":        "/path/to/private/key",
				"TRUSTED_SUBNET":    "10.0.0.0/8",
				"CONFIG":            "",
				"TRANSPORT":         "grpc",
			},
			config: "",
			want: Config{
				Addr:            "example.com:1234",
				StoreInterval:   100,
				FileStoragePath: "/tmp/metrics.dmp",
				Restore:         false,
				DSN:             "user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable",
				HashKey:         "secret",
				CryptoKeyPath:   "/path/to/private/key",
				TrustedSubnet:   "10.0.0.0/8",
				ConfigPath:      "",
				Transport:       "grpc",
			},
		},
		{
			testName: "Config.json",
			envs: map[string]string{
				"CONFIG": os.TempDir() + "config-slucigOmlasBigshAjBadOaweujdeng9.json",
			},
			config: `
				{
					"address": "example.com:8888",
					"store_interval": 200,
					"store_file": "/path/to/dump",
					"restore": true,
					"database_dsn": "user=myuser password=pswd host=10.0.0.1 port=6432 dbname=mydb sslmode=enable",
					"hash_key": "mykey",
					"crypto_key": "/path/to/key.pem",
					"trusted_subnet": "0.0.0.0/0",
					"transport": "grpc"
				}
			`,
			want: Config{
				Addr:            "example.com:8888",
				StoreInterval:   200,
				FileStoragePath: "/path/to/dump",
				Restore:         true,
				DSN:             "user=myuser password=pswd host=10.0.0.1 port=6432 dbname=mydb sslmode=enable",
				HashKey:         "mykey",
				CryptoKeyPath:   "/path/to/key.pem",
				TrustedSubnet:   "0.0.0.0/0",
				ConfigPath:      os.TempDir() + "config-slucigOmlasBigshAjBadOaweujdeng9.json",
				Transport:       "grpc",
			},
		},
		{
			testName: "Mixed",
			envs: map[string]string{
				"ADDRESS":        "example.com:65000",
				"RESTORE":        "false",
				"KEY":            "secret",
				"CRYPTO_KEY":     "/path/to/private/key",
				"TRUSTED_SUBNET": "10.1.0.0/16",
				"CONFIG":         os.TempDir() + "config-slucigOmlasBigshAjBadOaweujdeng9.json",
				"TRANSPORT":      "http",
			},
			config: `
				{
					"address": "example.com:8888",
					"restore": true,
					"database_dsn": "user=myuser password=pswd host=10.0.0.1 port=6432 dbname=mydb sslmode=enable",
					"hash_key": "mykey",
					"crypto_key": "/path/to/key.pem",
					"trusted_subnet": "192.168.1.0/24",
					"transport": "grpc"
				}
			`,
			want: Config{
				Addr:            "example.com:65000",
				StoreInterval:   300,
				FileStoragePath: "metrics.dmp",
				Restore:         false,
				DSN:             "user=myuser password=pswd host=10.0.0.1 port=6432 dbname=mydb sslmode=enable",
				HashKey:         "secret",
				CryptoKeyPath:   "/path/to/private/key",
				TrustedSubnet:   "10.1.0.0/16",
				ConfigPath:      os.TempDir() + "config-slucigOmlasBigshAjBadOaweujdeng9.json",
				Transport:       "http",
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
