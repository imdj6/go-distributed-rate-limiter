// internal/proxy/proxy.go
package proxy

import (
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(targetURL string) *httputil.ReverseProxy {
	// Parse the Node.js URL (e.g., "http://localhost:3000")
	target, _ := url.Parse(targetURL)

	// Go's standard library gives us a production-ready proxy out of the box
	return httputil.NewSingleHostReverseProxy(target)
}
