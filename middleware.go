package main

import (
	"net/http"
	"time"
)

func LoggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// before
		logger.Debugf(" In %v %v %v", r.RemoteAddr, r.Method, r.URL)

		// during
		next.ServeHTTP(w, r)

		// end
		// logger.Debugf("Out %v %v %v %v", r.RemoteAddr, r.Method, r.URL, w.Header())
		logger.Debugf("Out %v %v %v %v", r.RemoteAddr, r.Method, r.URL, time.Since(start))
	})
}

func SetHeadersMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}
