package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

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
		return fmt.Errorf("Canot parse %s. Must be host:port", flagValue)
	}
	n.Host = params[0]
	n.Port, err = strconv.Atoi(params[1])
	if err != nil {
		return err
	}
	return nil
}

func httpServer() {
	var addr netAddress
	var addr_env string = os.Getenv("ADDRESS")

	if addr_env != "" {
		if err := addr.Set(addr_env); err != nil {
			panic(err)
		}
	} else {
		addr = netAddress{Host: "localhost", Port: 8080}
		flag.Var(&addr, "a", "Listen address. Format host:port")
		flag.Parse()
	}

	r := chi.NewRouter()
	r.Get(`/`, getAllParams)
	r.Get(`/value/{kind}/{name}`, getParam)
	r.Post(`/update/{kind}/{name}/{value}`, setParam)

	err := http.ListenAndServe(addr.String(), r)
	if err != nil {
		panic(err)
	}
}
