// Package config is used to get command-line and Env Metrics-Server settings.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type (
	// Config is a Metrics-Server configuration storage.
	Config struct {
		Addr            string `env:"ADDRESS" json:"address"`
		StoreInterval   int    `env:"STORE_INTERVAL" json:"store_interval"`
		FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
		Restore         bool   `env:"RESTORE" json:"restore"`
		DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
		HashKey         string `env:"KEY" json:"hash_key"`
		CryptoKeyPath   string `env:"CRYPTO_KEY" json:"crypto_key"`
		ConfigPath      string `env:"CONFIG"`
	}
	netAddress struct {
		Host string
		Port int
	}
)

// String converts netAddress to string to implement flag.Value interface.
func (n *netAddress) String() string {
	return fmt.Sprint(n.Host, ":", n.Port)
}

// Set converts string to netAddress to implement flag.Value interface.
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

// Get is a single method to get all Metrics-Server settings.
func (cfg *Config) Get() error {

	// first: read config file with low priority
	fs := flag.NewFlagSet("metrics-server-preread", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigPath, "config", "", "Config file path (short: -c)")
	fs.StringVar(&cfg.ConfigPath, "c", "", "short for -config")
	fs.Parse(os.Args[1:])

	if config, ok := os.LookupEnv("CONFIG"); ok {
		cfg.ConfigPath = config
	}

	if cfg.ConfigPath != "" {
		if err := cfg.readFromFile(); err != nil {
			return fmt.Errorf("cannot read config file:%v", err)
		}
	}

	Addr := netAddress{Host: "localhost", Port: 8080}

	fs = flag.NewFlagSet("metrics-server", flag.ContinueOnError)

	fs.Var(&Addr, "a", "Listen address. Format host:port, default localhost:8080")
	StoreInterval := fs.Int("i", 300, "Store interval. Format int, default 300.")
	Restore := fs.Bool("r", true, "Restore data from disk on start. Format bool, default true.")
	FileStoragePath := fs.String("f", "metrics.dmp", "File to store data. Format string, default metrics.dmp.")
	DSN := fs.String("d", "", "PostrgeSQL DSN. Format: \"user=postgres password=secret host=localhost port=5432 dbname=mydb sslmode=disable\"")
	HashKey := fs.String("k", "", "Hash Key")
	CryptoKeyPath := fs.String("crypto-key", "", "Private Key path")
	fs.Parse(os.Args[1:])

	// replace config file options with explicit cmd-line values
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Addr = Addr.String()
		case "i":
			cfg.StoreInterval = *StoreInterval
		case "r":
			cfg.Restore = *Restore
		case "f":
			cfg.FileStoragePath = *FileStoragePath
		case "d":
			cfg.DSN = *DSN
		case "k":
			cfg.HashKey = *HashKey
		case "crypto-key":
			cfg.CryptoKeyPath = *CryptoKeyPath
		}
	})

	// replace only empty config values with cmd-line defaults
	if cfg.Addr == "" {
		cfg.Addr = Addr.String()
	}

	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = *StoreInterval
	}

	if cfg.Restore == false {
		cfg.Restore = *Restore
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *FileStoragePath
	}

	// read envs with high priority
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("cannot parse env: %v", err)
	}

	return nil
}

// readFromFile reads config from json changing only unspecified values
func (cfg *Config) readFromFile() error {
	cfgData, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("error reading file:%v", err)
	}

	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		return fmt.Errorf("error parsing json config:%v", err)
	}
	return nil
}
