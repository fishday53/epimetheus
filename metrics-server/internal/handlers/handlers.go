package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"metrics-server/internal/config"
	"metrics-server/internal/log"
	"metrics-server/internal/storage"
	"metrics-server/internal/storage/memory"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AppContext struct {
	Name string
	DB   storage.Repositories
	Log  *zap.SugaredLogger
	Cfg  *config.Config
}

func NewAppContext(name string, cfg *config.Config) *AppContext {
	return &AppContext{
		Name: name,
		DB:   memory.NewMemStorage(name),
		Log:  log.NewLogger(),
		Cfg:  cfg,
	}
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
		ctx.Log.Errorln("Name is not defined")
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
		ctx.Log.Errorln("Unsupported metric type")
		return
	}

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		ctx.Log.Errorln(err.Error())
		return
	}

	_, err = ctx.DB.Set(&metric)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if ctx.Cfg.StoreInterval == 0 {
		err = ctx.DB.Dump(ctx.Cfg.FileStoragePath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	res.WriteHeader(http.StatusOK)
}

func (ctx *AppContext) GetParam(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric
	var resultString string

	metric.ID = chi.URLParam(req, "name")

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		ctx.Log.Errorln("Name is not defined")
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
		ctx.Log.Errorln("Unsupported metric type")
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
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

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	for _, s := range *result {
		switch s.MType {
		case "gauge":
			resultString = storage.GaugeToString(*s.Value)
		case "counter":
			resultString = storage.CounterToString(*s.Delta)
		default:
			res.WriteHeader(http.StatusInternalServerError)
			ctx.Log.Errorln("Unsupported metric type")
			return
		}
		fmt.Fprintf(res, "%s:\t%s\n", s.ID, resultString)
	}
}

func (ctx *AppContext) SetParamJSON(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric

	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		ctx.Log.Errorln(err.Error())
		return
	}

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		ctx.Log.Errorln("Name is not defined")
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

	if ctx.Cfg.StoreInterval == 0 {
		err = ctx.DB.Dump(ctx.Cfg.FileStoragePath)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) GetParamJSON(res http.ResponseWriter, req *http.Request) {
	var metric storage.Metric

	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		ctx.Log.Errorln(err.Error())
		return
	}

	if metric.ID == "" {
		res.WriteHeader(http.StatusNotFound)
		ctx.Log.Errorln("Name is not defined")
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

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
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

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "%s", jsonData)
}

func (ctx *AppContext) CheckDBConnect(res http.ResponseWriter, req *http.Request) {

	if ctx.Cfg.DSN == "" {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("pgx", ctx.Cfg.DSN)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(c); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
