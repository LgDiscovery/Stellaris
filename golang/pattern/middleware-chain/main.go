package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("request", "method", r.Method, "url", r.URL.String(), "duration", time.Since(start))
		})
	}
}

func Auth(secret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != secret {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Chain(h http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("orders ok"))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", ordersHandler)
	handler := Chain(mux, Logging(logger), Auth("my-secret-token"))
	http.ListenAndServe(":8080", handler)
}
