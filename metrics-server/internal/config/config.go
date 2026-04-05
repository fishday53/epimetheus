package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type (
	Config struct {
		Addr            string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
		DSN             string `env:"DATABASE_DSN"`
		HashKey         string `env:"KEY"`
	}
	netAddress struct {
		Host string
		Port int
	}
)

func (n *netAddress) String() string {
	return fmt.Sprint(n.Host, ":", n.Port)
}

func (n *netAddress) Set(flagValue string) error {
	var err error
	params := strings.Split(flagValue, ":")
	if len(params) != 2 {
		return fmt.Errorf("cannot parse %s. must be host:port", flagValue)
	}
	n.Host = params[0]
	n.Port, err = strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("port definition error:%v", err)
	}
	return nil
}

func (cfg *Config) Get() error {
	addr := netAddress{Host: "localhost", Port: 8080}

	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("cannot parse env: %v", err)
	}

	fs := flag.NewFlagSet("metrics-server", flag.ContinueOnError)

	fs.Var(&addr, "a", "Listen address. Format host:port, default localhost:8080")
	storeIntervalFlag := fs.Int("i", 300, "Store interval. Format int, default 300.")
	restoreFlag := fs.Bool("r", true, "Restore data from disk on start. Format bool, default true.")
	fileStoragePathFlag := fs.String("f", "metrics.dmp", "File to store data. Format string, default metrics.dmp.")
	dsnFlag := fs.String("d", "", "PostrgeSQL DSN. Format: \"user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable\"")
	hashKey := fs.String("k", "", "Hash Key")

	fs.Parse(os.Args[1:])

	if cfg.Addr != "" {
		if err = addr.Set(cfg.Addr); err != nil {
			return fmt.Errorf("cannot set address: %v", err)
		}
	} else {
		cfg.Addr = addr.String()
	}

	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = *storeIntervalFlag
	}

	_, ok := os.LookupEnv("RESTORE")
	if !cfg.Restore && !ok {
		cfg.Restore = *restoreFlag
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *fileStoragePathFlag
	}

	if cfg.DSN == "" {
		cfg.DSN = *dsnFlag
	}

	if cfg.HashKey == "" {
		cfg.HashKey = *hashKey
	}

	return nil
}
