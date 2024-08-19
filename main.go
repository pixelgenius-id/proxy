package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Backend servers
var servers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

// Function to check server health
func isServerHealthy(serverURL string) bool {
	client := http.Client{
		Timeout: 2 * time.Second, // Set a short timeout for health checks
	}
	resp, err := client.Get(serverURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

// Function to select the best server based on availability
func getAvailableServer() string {
	for _, server := range servers {
		if isServerHealthy(server) {
			return server
		}
	}
	return "" // Return empty string if no servers are available
}

// Reverse proxy handler
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	backendURL := getAvailableServer()

	if backendURL == "" {
		http.Error(w, "No healthy backend servers available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Forwarding request to: %s", backendURL)

	parsedURL, err := url.Parse(backendURL)
	if err != nil {
		http.Error(w, "Invalid backend server URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ServeHTTP(w, r)
}

// Health check handler for the proxy itself
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Proxy is running\n")
}

func main() {
	// Set up routes
	http.HandleFunc("/", proxyHandler)
	http.HandleFunc("/health", healthCheckHandler)

	// Start the proxy server
	log.Println("Starting proxy server on port 3040...")
	err := http.ListenAndServe(":3040", nil)
	if err != nil {
		log.Fatalf("Could not start proxy server: %v", err)
	}
}
