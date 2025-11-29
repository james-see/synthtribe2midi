// Package main is the entry point for the synthtribe2midi API server
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/james-see/synthtribe2midi/pkg/api"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	fmt.Printf("Starting synthtribe2midi API server on port %d...\n", *port)
	fmt.Printf("Swagger docs available at http://localhost:%d/swagger/index.html\n", *port)
	
	if err := api.StartServer(*port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

