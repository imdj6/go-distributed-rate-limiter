package main

import (
	"fmt"
	"log"
	"my-gateway/internal/config"
	"my-gateway/internal/limiter"
	"net/http"
)

func main() {
	fmt.Println("🚀 Booting up API Gateway...")

	store, err := config.NewRedisStore()

	if err != nil {
		log.Fatal("Critical boot failure: %v", err)
	}

	defer store.Client.Close()

	log.Println("✅ Successfully connected to Redis!")

	rateLimiter := limiter.NewRateLimiterService(store)

	log.Println("🛡️ Rate Limiter Engine Online")

	// 2. Define a simple Route
	// http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	// Get the user's IP address
	// 	userIP := r.RemoteAddr

	// 	// Print to our terminal
	// 	fmt.Printf("Incoming request from: %s\n", userIP)

	// 	// Send response to the browser
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte("Gateway is alive!"))
	// })

	// secretHandler := func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(`{"data": "Here is the top-secret database payload!"}`))
	// }

	http.HandleFunc("/health", rateLimiter.RateLimitMiddleware())

	// 3. Start the Server
	port := ":8080"
	log.Printf("🌐 Gateway server listening on http://localhost%s\n", port)

	// ListenAndServe blocks the main thread and keeps the app running forever
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
