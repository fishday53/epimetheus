package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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
		return fmt.Errorf("canot parse %s. must be host:port", flagValue)
	}
	n.Host = params[0]
	n.Port, err = strconv.Atoi(params[1])
	if err != nil {
		return err
	}
	return nil
}

func (cfg *Config) Get() error {
	var addr netAddress
	var addrEnv = os.Getenv("ADDRESS")
	var storeInterval = os.Getenv("STORE_INTERVAL")
	var restore = os.Getenv("RESTORE")
	var err error

	if addrEnv != "" {
		if err = addr.Set(addrEnv); err != nil {
			return err
		}
	} else {
		addr = netAddress{Host: "localhost", Port: 8080}
		flag.Var(&addr, "a", "Listen address. Format host:port, default localhost:8080")
	}

	if storeInterval != "" {
		cfg.StoreInterval, err = strconv.Atoi(storeInterval)
		if err != nil {
			return err
		}
	} else {
		cfg.StoreInterval = *flag.Int("i", 300, "Store interval. Format int, default 300.")
	}

	cfg.FileStoragePath = os.Getenv("FILE_STORAGE_PATH")

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *flag.String("f", "metrics.dmp", "File to store data. Format string, default metrics.dmp.")
	}

	if restore != "" {
		cfg.Restore, err = strconv.ParseBool(restore)
		if err != nil {
			return err
		}
	} else {
		cfg.Restore = *flag.Bool("r", true, "Restore data from disk on start. Format bool, default true.")
	}

	flag.Parse()

	cfg.Addr = addr.String()

	return nil
}
