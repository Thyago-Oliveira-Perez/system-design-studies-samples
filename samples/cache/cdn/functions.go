package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

/*
owner → read/write/execute
others → read/execute
*/
const permissionCode = 0755
const cacheFolder = "./cache"
const origin = "https://www.uol.com.br/"

func handleCacheFolderCreation() {
	err := os.MkdirAll(cacheFolder, permissionCode)
	if err != nil {
		panic(err)
	}
}

func handleHealthResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "health")
}

func handleHashCreation(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func handleContentResult(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	cachePath := filepath.Join(cacheFolder, handleHashCreation(fmt.Sprintf("%s%s%s", r.Method, r.URL.Path, r.URL.RawQuery)))
	
	var body []byte

	_, err := os.Stat(cachePath)

	if os.IsNotExist(err) {
		log.Println("Cache miss")

		url := fmt.Sprintf("%s%s", origin, r.URL.Path)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Error fetching from origin", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading origin response", http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(cachePath, body, 0644)
		if err != nil {
			log.Printf("Failed to write cache: %v", err)
		}

	} else if err == nil {
		log.Printf("Cache hit: %s", cachePath)

		// Read from cache
		body, err = os.ReadFile(cachePath)
		if err != nil {
			http.Error(w, "Error reading cache", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Error checking cache", http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("Request %s %s took %s\n", r.Method , r.URL, elapsed)
	w.Write(body)
}