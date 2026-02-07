package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"metrics-server/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AppContext struct {
	DB  storage.Repositories
	Log *zap.SugaredLogger
}

func (ctx *AppContext) SetParam(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric
	var err error

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metric.ID = chi.URLParam(req, "name")

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		log.Println("Name is not defined")
		return
	}

	metric.MType = chi.URLParam(req, "mtype")

	switch metric.MType {
	case "gauge":
		metric.Value, err = storage.StringToGauge(chi.URLParam(req, "value"))
	case "counter":
		metric.Delta, err = storage.StringToCounter(chi.URLParam(req, "value"))
	default:
		res.WriteHeader(http.StatusBadRequest)
		log.Println("Unsupported metric type")
		return
	}

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	_, err = ctx.DB.Set(&metric)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (ctx *AppContext) GetParam(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric
	var resultString string

	metric.ID = chi.URLParam(req, "name")

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		log.Println("Name is not defined")
		return
	}

	metric.MType = chi.URLParam(req, "mtype")

	result, err := ctx.DB.Get(&metric)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(res, "Value of %s is absent\n", metric.ID)
		return
	}

	switch metric.MType {
	case "gauge":
		resultString = storage.GaugeToString(*result.Value)
	case "counter":
		resultString = storage.CounterToString(*result.Delta)
	default:
		res.WriteHeader(http.StatusBadRequest)
		log.Println("Unsupported metric type")
		return
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s\n", resultString)
}

func (ctx *AppContext) GetAllParams(res http.ResponseWriter, req *http.Request) {
	var resultString string

	result, err := ctx.DB.GetAll()
	if err != nil {
		res.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(res, "Something went wrong\n")
		return
	}

	res.WriteHeader(http.StatusOK)
	for _, s := range *result {
		switch s.MType {
		case "gauge":
			resultString = storage.GaugeToString(*s.Value)
		case "counter":
			resultString = storage.CounterToString(*s.Delta)
		default:
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Unsupported metric type")
			return
		}
		fmt.Fprintf(res, "%s:\t%s\n", s.ID, resultString)
	}
}

func (ctx *AppContext) SetParamJSON(res http.ResponseWriter, req *http.Request) {
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
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) GetParamJSON(res http.ResponseWriter, req *http.Request) {
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
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) GetAllParamsJSON(res http.ResponseWriter, req *http.Request) {
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
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(res, "%s", jsonData)
}
