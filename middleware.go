package limiter

import (
	"fmt"
	"my-gateway/internal/proxy"
	"net/http"
)

func (rl *RateLimiterService) RateLimitMiddleware() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Identify the user
		userIP := r.RemoteAddr

		// 2. Check the limit using the service we just built
		isAllowed, tokensRemaining, err := rl.CheckLimit(userIP)
		if err != nil {
			fmt.Printf("⚠️ Rate Limiter Warning: %v\n", err)
		}

		// 3. Set the standard REST headers
		w.Header().Set("X-RateLimit-Limit", "5")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", tokensRemaining))

		// 4. The Decision
		if !isAllowed {
			// They are blocked! We write the error and 'return'.
			// By returning here, we NEVER call the 'next' function. The request dies here.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too Many Requests", "message": "Bucket empty. Chill out!"}`))

			fmt.Printf("❌ BLOCKED: %s\n", userIP)
			return
		}

		// 5. Success! Call the actual route handler.
		// This is the exact equivalent of calling 'next()' in Node.js!
		fmt.Printf("✅ ALLOWED: %s (Tokens left: %d)\n", userIP, tokensRemaining)
		nodeJsProxy := proxy.NewReverseProxy("http://localhost:3000")
		nodeJsProxy.ServeHTTP(w, r)
	}
}
