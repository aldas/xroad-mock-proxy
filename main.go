package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	targetScheme := "https"
	targetHost := "www.google.com"

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: targetScheme,
		Host:   targetHost,
	})

	proxy.FlushInterval = 1 * time.Second

	proxy.Director = func(r *http.Request) {
		// To rewrite Host headers, we need to use ReverseProxy with a custom Director policy.
		r.Host = targetHost
		r.URL.Host = r.Host
		r.URL.Scheme = targetScheme
	}

	http.Handle("/", proxy)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
