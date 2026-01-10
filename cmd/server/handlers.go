package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// func mainPage(res http.ResponseWriter, req *http.Request) {
// 	res.WriteHeader(http.StatusNotFound)
// }

func setParam(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	if name == "" {
		res.WriteHeader(http.StatusNotFound)
		fmt.Println("Name is not defined")
		return
	}

	err := setMetric(Storage, kind, name, value)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func getParam(res http.ResponseWriter, req *http.Request) {
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	if result, err := getMetric(Storage, kind, name); err == nil {
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s\n", result)
	} else {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(res, "Value of %s is absent\n", name)
	}
}

func getAllParams(res http.ResponseWriter, req *http.Request) {
	if result, err := getAllMetrics(Storage); err == nil {
		res.WriteHeader(http.StatusOK)
		for _, s := range result {
			fmt.Fprintf(res, "%s", s)
		}
	} else {
		res.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(res, "Something went wrong\n")
	}
}
