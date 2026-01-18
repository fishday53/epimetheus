package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr string `env:"ADDRESS"`
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

	if addrEnv != "" {
		if err := addr.Set(addrEnv); err != nil {
			return err
		}
	} else {
		addr = netAddress{Host: "localhost", Port: 8080}
		flag.Var(&addr, "a", "Listen address. Format host:port")
		flag.Parse()
	}

	cfg.Addr = addr.String()

	return nil
}
