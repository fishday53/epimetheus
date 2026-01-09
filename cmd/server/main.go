package main

import (
	"net/http"
)

type gauge float64
type counter int64

var Storage MemStorage

// curl -XPOST http://localhost:8080/update/gauge/a/1.53
// curl http://localhost:8080/show/a
// curl -XPOST http://localhost:8080/update/counter/b/-1
// curl http://localhost:8080/show/b
func main() {

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)
	mux.HandleFunc(`/update/{kind}/{name}/{value}`, setParam)
	mux.HandleFunc(`/show/{name}`, getParam)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		panic(err)
	}
}
