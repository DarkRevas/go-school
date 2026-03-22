package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.ListenAndServe(":8080", mux)
}

// Подними HTTP-сервер на :8080 и реализуй маршрут:

// GET /health -> 200 OK, тело ok.

// Требование: используй http.NewServeMux() и шаблон маршрута с методом.