package handlers

import (
	"fmt"
	"log"
	"metrics-server/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AppContext struct {
	DB storage.Repositories
}

func (ctx *AppContext) SetParam(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	if name == "" {
		res.WriteHeader(http.StatusNotFound)
		log.Println("Name is not defined")
		return
	}

	err := ctx.DB.Set(kind, name, value)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (ctx *AppContext) GetParam(res http.ResponseWriter, req *http.Request) {
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")

	result, err := ctx.DB.Get(kind, name)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(res, "Value of %s is absent\n", name)
		return
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s\n", result)
}

func (ctx *AppContext) GetAllParams(res http.ResponseWriter, req *http.Request) {
	result, err := ctx.DB.GetAll()
	if err != nil {
		res.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(res, "Something went wrong\n")
		return
	}

	res.WriteHeader(http.StatusOK)
	for _, s := range result {
		fmt.Fprintf(res, "%s", s)
	}
}
