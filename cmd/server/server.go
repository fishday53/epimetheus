package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func httpServer() {

	r := chi.NewRouter()
	r.Get(`/`, getAllParams)
	r.Get(`/value/{kind}/{name}`, getParam)
	r.Post(`/update/{kind}/{name}/{value}`, setParam)

	err := http.ListenAndServe(`localhost:8080`, r)
	if err != nil {
		panic(err)
	}
}
