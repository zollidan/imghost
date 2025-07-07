package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/file/upload", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("hello")
	})


	fmt.Println("Server is running on http://localhost:8000/")
	http.ListenAndServe("localhost:8000", r)
}