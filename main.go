package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

/*
	Roadmap
	1.IPFS
	2.AES-256-GCM (https://dev.to/breda/secret-key-encryption-with-go-using-aes-316d)
	3.Infura or Pinata
	4.ui
*/

var 
(
    tmpl *template.Template
    host string = "localhost:8000"
)

type IndexData struct {
	Title   string
	Header  string
	Message string
	IsError bool
}

type EncryptionResult struct {
    EncryptedData string `json:"encrypted_data"`
    Nonce         string `json:"nonce"`
    Key           string `json:"key"`
}



func init(){
	tmpl = template.Must(template.ParseFiles("index.html"))

	os.MkdirAll("./uploads", 0755)
	os.MkdirAll("./encrypted", 0755)
}

// создает случайный 32-байтовый ключ
func generateRandomKey() ([]byte, error) {
    key := make([]byte, 32)
    _, err := rand.Read(key)
    return key, err
}

// encryptData шифрует данные с использованием AES-256-GCM
func encryptData(data []byte, key []byte) ([]byte, []byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, nil, err
    }

    ciphertext := gcm.Seal(nil, nonce, data, nil)
    
    return ciphertext, nonce, nil
}

func encryptFile(inputPath, outputPath string, key []byte) ([]byte, error) {
    data, err := os.ReadFile(inputPath)
    if err != nil {
        return nil, err
    }

    ciphertext, nonce, err := encryptData(data, key)
    if err != nil {
        return nil, err
    }

    encryptedFile := struct {
        Data  []byte `json:"data"`
        Nonce []byte `json:"nonce"`
    }{
        Data:  ciphertext,
        Nonce: nonce,
    }

    jsonData, err := json.Marshal(encryptedFile)
    if err != nil {
        return nil, err
    }

    err = os.WriteFile(outputPath, jsonData, 0644)
    if err != nil {
        return nil, err
    }

    return nonce, nil
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

    r.Post("/file/upload", func(w http.ResponseWriter, r *http.Request) {
        // Парсим форму (10 MB max)
        err := r.ParseMultipartForm(10 << 20)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to parse multipart form: %v", err), http.StatusBadRequest)
            return
        }

        file, fileHeader, err := r.FormFile("file")
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
            return
        }
        defer file.Close()
        
        var key []byte
 
		key, err = generateRandomKey()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate key: %v", err), http.StatusInternalServerError)
			return
		}
        
        originalPath := filepath.Join("./uploads", fileHeader.Filename)
        dst, err := os.Create(originalPath)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to create file on server: %v", err), http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            http.Error(w, fmt.Sprintf("Failed to copy file content: %v", err), http.StatusInternalServerError)
            return
        }

        encryptedPath := filepath.Join("./encrypted", fileHeader.Filename+".enc")
        nonce, err := encryptFile(originalPath, encryptedPath, key)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to encrypt file: %v", err), http.StatusInternalServerError)
            return
        }

        os.Remove(originalPath)

        response := map[string]interface{}{
            "message":        "File uploaded and encrypted successfully!",
            "filename":       fileHeader.Filename,
            "encrypted_file": fileHeader.Filename + ".enc",
            "size":          fmt.Sprintf("%d bytes", fileHeader.Size),
            "key":           base64.StdEncoding.EncodeToString(key),
            "nonce":         base64.StdEncoding.EncodeToString(nonce),
        }

        w.Header().Add("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })

	fmt.Printf("Server is running on http://%s/", host)
	http.ListenAndServe(host, r)
}