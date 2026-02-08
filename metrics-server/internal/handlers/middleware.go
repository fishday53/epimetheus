package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	gzipWriter struct {
		http.ResponseWriter
		Writer io.Writer
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func (ctx *AppContext) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		ctx.Log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	})
}

func (ctx *AppContext) CheckContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid Content-Type, expected application/json", http.StatusUnsupportedMediaType)
			ctx.Log.Errorln("Invalid Content-Type ", r.Header.Get("Content-Type"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ctx *AppContext) GzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("we're here")
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Bad Request: Invalid gzip data", http.StatusBadRequest)
				ctx.Log.Errorln("Bad Request: Invalid gzip data")
				return
			}
			defer gzr.Close()
			//ctx.Log.Infoln("gzip from client")
			//fmt.Println("gzip from client")
			// originalBody := r.Body
			// defer originalBody.Close()
			// fmt.Println("bedore:", r.Body)
			r.Body = gzr
			// fmt.Println("adfter:", r.Body)
			// r.ContentLength = -1
			// r.Header.Del("Content-Encoding")
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			//ctx.Log.Infoln("no gzip to client")
			//fmt.Println("no gzip to client")
			next.ServeHTTP(w, r)
			return
		}

		gzw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gzw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		//ctx.Log.Infoln("gzip to client")
		fmt.Println("gzip to client")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzw}, r)
	})
}
