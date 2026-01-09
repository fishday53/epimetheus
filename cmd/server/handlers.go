package main

import (
	"fmt"
	"net/http"
)

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
		fmt.Println("Name is not defined")
		return
	}

	err := UpdateMetric(&Storage, kind, name, value) // Storage.Update(kind, name, value)
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
