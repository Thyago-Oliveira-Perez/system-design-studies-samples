package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 8080

func main() {

	handleCacheFolderCreation()

	http.HandleFunc("/health", handleHealthResult)

	http.HandleFunc("/", handleContentResult)

	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}