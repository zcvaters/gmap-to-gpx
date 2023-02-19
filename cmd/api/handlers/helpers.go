package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WithStatus(code int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		h.ServeHTTP(w, r)
	})
}

func Error(err error, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), code)
	})
}

func JSON(v interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(v)
	})
}

func Text(v interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, v)
	})
}
