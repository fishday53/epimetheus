package config

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DSN             string `env:"DATABASE_DSN"`
}

type netAddress struct {
	Host string
	Port int
}

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

	flag.Var(&addr, "a", "Listen address. Format host:port, default localhost:8080")
	storeIntervalFlag := flag.Int("i", 300, "Store interval. Format int, default 300.")
	restoreFlag := flag.Bool("r", true, "Restore data from disk on start. Format bool, default true.")
	fileStoragePathFlag := flag.String("f", "metrics.dmp", "File to store data. Format string, default metrics.dmp.")
	dsnFlag := flag.String("d", "", "PostrgeSQL DSN. Format: \"user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable\"")

	flag.Parse()

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

	if !cfg.Restore {
		cfg.Restore = *restoreFlag
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *fileStoragePathFlag
	}

	if cfg.DSN == "" {
		cfg.DSN = *dsnFlag
	}

	return nil
}
