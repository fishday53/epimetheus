package handlers

import (
	"encoding/json"
	"fmt"
	"metrics-server/internal/storage"
	"metrics-server/internal/usecase"
	"metrics-server/internal/usecase/context"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
)

func SetParam(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metric usecase.Metric
		var err error

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		metric.ID = chi.URLParam(req, "name")

		if metric.ID == "" {
			res.WriteHeader(http.StatusNotFound)
			app.Log.Errorln("Name is not defined")
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
			app.Log.Errorln("Unsupported metric type")
			return
		}

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			app.Log.Errorln(err.Error())
			return
		}

		_, err = app.DB.Set(&metric)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if app.Cfg.StoreInterval == 0 {
			err = app.DB.Dump(app.Cfg.FileStoragePath)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func GetParam(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metric usecase.Metric
		var resultString string

		metric.ID = chi.URLParam(req, "name")

		if metric.ID == "" {
			res.WriteHeader(http.StatusNotFound)
			app.Log.Errorln("Name is not defined")
			return
		}

		metric.MType = chi.URLParam(req, "mtype")

		result, err := app.DB.Get(&metric)
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
			app.Log.Errorln("Unsupported metric type")
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s\n", resultString)
	}
}

func GetAllParams(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var resultString string

		result, err := app.DB.GetAll()
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
				app.Log.Errorln("Unsupported metric type")
				return
			}
			fmt.Fprintf(res, "%s:\t%s\n", s.ID, resultString)
		}
	}
}

func SetParamJSON(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metric usecase.Metric

		if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			app.Log.Errorln("Cannot decode request:", err)
			return
		}

		if metric.ID == "" {
			res.WriteHeader(http.StatusNotFound)
			app.Log.Errorln("Name is not defined")
			return
		}

		result, err := app.DB.Set(&metric)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			app.Log.Errorln("Cannot set metric:", err)
			return
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "Error in marshaler: %v\n", err)
			app.Log.Errorln("Error in marshaller:", err)
			return
		}

		if app.Cfg.StoreInterval == 0 {
			err = app.DB.Dump(app.Cfg.FileStoragePath)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				app.Log.Errorln("Dump error:", err)
				return
			}
		}

		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s", jsonData)
	}
}

func SetMultiParamJSON(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metrics []usecase.Metric

		if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			app.Log.Errorln("Cannot decode request:", err)
			return
		}

		for _, metric := range metrics {
			fmt.Println("metric:", metric.ID)
			if metric.ID == "" {
				res.WriteHeader(http.StatusNotFound)
				app.Log.Errorln("Name is not defined")
				return
			}
			_, err := app.DB.Set(&metric)
			if err != nil {
				res.WriteHeader(http.StatusBadRequest)
				app.Log.Errorln("Cannot set metric:", err)
				return
			}
		}

		if app.Cfg.StoreInterval == 0 {
			err := app.DB.Dump(app.Cfg.FileStoragePath)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				app.Log.Errorln("Dump error:", err)
				return
			}
		}
		jsonData, err := json.Marshal(&metrics)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			app.Log.Errorln("Multianswer error:", err)
			return
		}
		//fmt.Println("SetMultiParamJSON4:", jsonData)
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s", jsonData)
	}
}

func GetParamJSON(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metric usecase.Metric

		if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			app.Log.Errorln(err.Error())
			return
		}

		if metric.ID == "" {
			res.WriteHeader(http.StatusNotFound)
			app.Log.Errorln("Name is not defined")
			return
		}

		result, err := app.DB.Get(&metric)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(res, "Value of %s is absent\n", metric.ID)
			app.Log.Errorln("Cannot get metric:", err)
			return
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "Error in marshaller: %v\n", err)
			app.Log.Errorln("Error in marshaller:", err)
			return
		}

		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s", jsonData)
	}
}

func GetAllParamsJSON(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		result, err := app.DB.GetAll()
		if err != nil {
			res.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(res, "Something went wrong\n")
			app.Log.Errorln("Cannot get all metrics:", err)
			return
		}

		jsonData, err := json.Marshal(*result)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "Error in marshaller: %v\n", err)
			app.Log.Errorln("Error in marchaller:", err)
			return
		}

		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%s", jsonData)
	}
}

func CheckDBConnect(app *context.AppContext) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if app.Cfg.DSN == "" {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := app.DB.Ping(); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			app.Log.Errorln("DB check test failed:", err)
		}

		res.WriteHeader(http.StatusOK)
	}
}
