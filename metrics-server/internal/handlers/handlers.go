package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"metrics-server/internal/storage"
	"net/http"

	"go.uber.org/zap"
)

type AppContext struct {
	DB  storage.Repositories
	Log *zap.SugaredLogger
}

func (ctx *AppContext) SetParam(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric

	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		log.Println("Name is not defined")
		return
	}

	result, err := ctx.DB.Set(&metric)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "Error in marshaler: %v\n", err)
		return
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) GetParam(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric

	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		log.Println("Name is not defined")
		return
	}

	result, err := ctx.DB.Get(&metric)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(res, "Value of %s is absent\n", metric.ID)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "Error in marshaler: %v\n", err)
		return
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) GetAllParams(res http.ResponseWriter, req *http.Request) {
	result, err := ctx.DB.GetAll()
	if err != nil {
		res.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(res, "Something went wrong\n")
		return
	}

	jsonData, err := json.Marshal(*result)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "Error in marshaler: %v\n", err)
		return
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s", jsonData)
}
