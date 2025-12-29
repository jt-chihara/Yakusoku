package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/jt-chihara/yakusoku/internal/broker"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	storage := broker.NewMemoryStorage()
	api := broker.NewAPI(storage)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Yakusoku Broker starting on %s\n", addr)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /pacts                                                    - List all contracts")
	fmt.Println("  GET  /pacts/provider/{provider}                                - Get contracts by provider")
	fmt.Println("  GET  /pacts/provider/{provider}/consumer/{consumer}/version/{version} - Get specific contract")
	fmt.Println("  GET  /pacts/provider/{provider}/consumer/{consumer}/latest     - Get latest contract")
	fmt.Println("  POST /pacts/provider/{provider}/consumer/{consumer}/version/{version} - Publish contract")
	fmt.Println("  GET  /matrix                                                   - Can I deploy check")

	if err := http.ListenAndServe(addr, api.Handler()); err != nil {
		log.Fatal(err)
	}
}
