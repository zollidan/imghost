package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

/*
	Roadmap
	1.IPFS
	2.Infura or Pinata
	3.Crypt
	4.ui
*/

type IndexData struct {
	Title   string
	Header  string
	Message string
	IsError bool
}

var tmpl *template.Template

func init(){
	tmpl = template.Must(template.ParseFiles("index.html"))
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := IndexData{
			Title:  "Загрузка файла",
			Header: "Загрузить ваш файл",
		}
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при рендеринге шаблона: %v", err), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/file/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")

		json.NewEncoder(w).Encode(filename)
	})

	r.Post("/file/upload", func(w http.ResponseWriter, r *http.Request) {
		// Оператор << в Go выполняет побитовый сдвиг влево (выделяю место под файл 10мб)
		// Set a maximum memory for parsing the form data (e.g., 10 MB)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
			return
		}
		// Get the file from the request. The "file" here refers to the name
		// attribute in your HTML input type="file" tag.
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a new file on the server to save the uploaded content
		dst, err := os.Create("./uploads/" + fileHeader.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create file on server: %v", err), http.StatusInternalServerError)
			return
		}

		// Copy the uploaded file content to the new file on the server
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, fmt.Sprintf("Failed to copy file content: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message":  "File uploaded successfully!",
			"filename": fileHeader.Filename,
			"size":     fmt.Sprintf("%d bytes", fileHeader.Size),
		}
		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response) 
	})


	fmt.Println("Server is running on http://localhost:8000/")
	http.ListenAndServe("localhost:8000", r)
}