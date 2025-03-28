package main

import (
	"fmt"
	"net/http"
)

var middlewares = []func(http.Handler) http.Handler{
	LoggingMiddleware,
}

func initMiddleware() {
	for _, m := range middlewares {
		server.router.Use(m)
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %v\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})
}
