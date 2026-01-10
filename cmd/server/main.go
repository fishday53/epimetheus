package main

// type gauge float64
// type counter int64

var Storage *memStorage

// curl -XPOST http://localhost:8080/update/gauge/a/1.53
// curl http://localhost:8080/value/gauge/a
// curl -XPOST http://localhost:8080/update/counter/b/-1
// curl http://localhost:8080/value/counter/b
func main() {
	Storage = NewMemStorage()

	httpServer()
}
