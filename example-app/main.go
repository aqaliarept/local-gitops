package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Readiness check endpoint
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Ready")
	})

	// Main application endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Local GitOps Example App</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        .info { background: #e8f4fd; padding: 15px; border-radius: 4px; margin: 20px 0; }
        .status { color: #28a745; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Local GitOps Example App</h1>
        <div class="info">
            <p><strong>Status:</strong> <span class="status">Running</span></p>
            <p><strong>Environment:</strong> %s</p>
            <p><strong>Time:</strong> %s</p>
            <p><strong>Port:</strong> %s</p>
        </div>
        <h2>Available Endpoints:</h2>
        <ul>
            <li><a href="/health">/health</a> - Health check</li>
            <li><a href="/ready">/ready</a> - Readiness check</li>
            <li><a href="/">/</a> - This page</li>
        </ul>
        <p>This is a simple Go HTTP server running in your local GitOps environment!</p>
    </div>
</body>
</html>`,
			os.Getenv("ENV"),
			time.Now().Format("2006-01-02 15:04:05"),
			port)
	})

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
