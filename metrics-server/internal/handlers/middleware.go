// Package handlers part is used to implement http-server middleware.
package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"metrics-server/internal/crypt"
	"metrics-server/internal/usecase/context"
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
	hashWriter struct {
		http.ResponseWriter
		Body   *bytes.Buffer
		Status int
	}
)

// Write is used to log responce size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader is used to log responce status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Write is used to compress the responce.
func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

// Write is used to calculate a hash.
func (w *hashWriter) Write(b []byte) (int, error) {
	return w.Body.Write(b)
}

// WriteHeader is used to write a hash header.
func (w *hashWriter) WriteHeader(statusCode int) {
	w.Status = statusCode
}

func getHash(hashKey string, b []byte) string {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(b[:])
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes[:])
}

// Logger is responcible for Metrics-Server log.
func Logger(app *context.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			app.Log.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)
		})
	}
}

// CheckContentType is used to check Content-Type headres for json-serving handlers.
func CheckContentType(app *context.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Invalid Content-Type, expected application/json", http.StatusUnsupportedMediaType)
				app.Log.Errorln("Invalid Content-Type ", r.Header.Get("Content-Type"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GzipHandler compresses responces.
func GzipHandler(app *context.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzr, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Bad Request: Invalid gzip data", http.StatusBadRequest)
					app.Log.Errorln("Bad Request: Invalid gzip data")
					return
				}
				defer gzr.Close()

				r.Body = gzr
			}

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
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

			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzw}, r)
		})
	}
}

// HashHandler is used to sign responces with hash header using HashKey secret.
func HashHandler(app *context.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if app.Cfg.HashKey == "" || r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			clientHash := r.Header.Get("Hashsha256")
			currentHash := getHash(app.Cfg.HashKey, bodyBytes)

			if clientHash != currentHash {
				http.Error(w, "bad body sign", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			wrapper := &hashWriter{
				ResponseWriter: w,
				Body:           &bytes.Buffer{},
				Status:         http.StatusOK,
			}

			next.ServeHTTP(wrapper, r)

			hash := getHash(app.Cfg.HashKey, wrapper.Body.Bytes())
			w.Header().Set("Hashsha256", hash)

			w.WriteHeader(wrapper.Status)
			w.Write(wrapper.Body.Bytes())
		})
	}
}

func CryptHandler(app *context.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if app.PrivKey == nil || r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			ciphertext, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			plaintext, err := crypt.Decrypt(app.PrivKey, ciphertext)
			if err != nil {
				http.Error(w, "Decryption failed", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(plaintext))
			next.ServeHTTP(w, r)
		})
	}
}
