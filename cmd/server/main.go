package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type gauge float64
type counter int64

type MemStorage struct {
	Value map[string]interface{}
}

var Storage MemStorage

func (m *MemStorage) Update(kind, name, value string) error {
	var err error

	if m.Value == nil {
		m.Value = make(map[string]interface{})
	}

	switch kind {
	case "gauge":
		m.Value[name], err = strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.New("Error gauge conversion")
		}
		fmt.Printf("gauge %s=%v\n", name, m.Value[name])
	case "counter":
		if _, ok := Storage.Value[name]; !ok {
			Storage.Value[name] = counter(0)
		}
		addition, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.New("Error counter conversion")
		}
		m.Value[name] = m.Value[name].(counter) + counter(addition)
		fmt.Printf("cntr %s=%v\n", name, m.Value[name])
	default:
		return errors.New("Unsupported value kind")
	}
	return nil
}

func mainPage(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNotFound)
}

func setParam(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	kind := req.PathValue("kind")
	name := req.PathValue("name")
	value := req.PathValue("value")

	if name == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	err := Storage.Update(kind, name, value)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func getParam(res http.ResponseWriter, req *http.Request) {
	name := req.PathValue("name")
	if result, ok := Storage.Value[name]; ok {
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "Value of %s is %v\n", name, result)
	} else {
		res.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(res, "Value of %s is absent\n", name)
	}
	//res.Write([]byte(fmt.Sprintf("Value of %s is %v", name, Storage.Value[name])))
}

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
