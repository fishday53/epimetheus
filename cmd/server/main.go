package main

var Storage *memStorage

func main() {
	Storage = NewMemStorage()

	httpServer()
}
