package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

/*
	Roadmap
	1.IPFS
	2.AES-256-GCM (https://dev.to/breda/secret-key-encryption-with-go-using-aes-316d)
	3.Infura or Pinata
	4.ui
*/

var (
	tmpl        *template.Template
	host        string = "localhost:8000"
	minioClient *minio.Client
	bucketName  string = "imghost-files"
)

type IndexData struct {
	Title   string
	Header  string
	Message string
	IsError bool
	Files   []EncryptedFile
}

type EncryptedFile struct {
	Name      string
	ExpiresAt time.Time
}

type EncryptionResult struct {
	EncryptedData string `json:"encrypted_data"`
	Nonce         string `json:"nonce"`
	Key           string `json:"key"`
}

func init() {
	tmpl = template.Must(template.ParseFiles("index.html"))

	os.MkdirAll("./uploads", 0755)
	os.MkdirAll("./encrypted", 0755)

	initMinIO()
}

func initMinIO() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin123"
	}

	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if bucketEnv := os.Getenv("MINIO_BUCKET"); bucketEnv != "" {
		bucketName = bucketEnv
	}

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal("Failed to initialize MinIO client:", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatal("Failed to check bucket existence:", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatal("Failed to create bucket:", err)
		}
		log.Printf("Bucket %s created successfully", bucketName)
	}
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

func encryptFile(inputPath, objectName string, key []byte) ([]byte, error) {
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

	// Upload to MinIO
	ctx := context.Background()
	reader := bytes.NewReader(jsonData)
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, int64(len(jsonData)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func scheduleDeletion(objectName string, d time.Duration) {
	time.AfterFunc(d, func() {
		ctx := context.Background()
		err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("Failed to delete object %s: %v", objectName, err)
		}
	})
}

// decryptData расшифровывает данные используя AES-256-GCM
func decryptData(ciphertext, nonce, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func decryptFile(encPath string, key []byte) ([]byte, error) {
	data, err := os.ReadFile(encPath)
	if err != nil {
		return nil, err
	}
	var ef struct {
		Data  []byte `json:"data"`
		Nonce []byte `json:"nonce"`
	}
	if err := json.Unmarshal(data, &ef); err != nil {
		return nil, err
	}
	return decryptData(ef.Data, ef.Nonce, key)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		entries, _ := os.ReadDir("./encrypted")
		files := make([]EncryptedFile, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, EncryptedFile{
				Name:      e.Name(),
				ExpiresAt: info.ModTime().Add(15 * time.Minute),
			})
		}
		data := IndexData{
			Title:  "Загрузка файла",
			Header: "Загрузить ваш файл",
			Files:  files,
		}
		if err := tmpl.Execute(w, data); err != nil {
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
		scheduleDeletion(encryptedPath, 15*time.Minute)

		response := map[string]interface{}{
			"message":        "File uploaded and encrypted successfully!",
			"filename":       fileHeader.Filename,
			"encrypted_file": fileHeader.Filename + ".enc",
			"size":           fmt.Sprintf("%d bytes", fileHeader.Size),
			"key":            base64.StdEncoding.EncodeToString(key),
			"nonce":          base64.StdEncoding.EncodeToString(nonce),
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	r.Post("/file/decrypt", func(w http.ResponseWriter, r *http.Request) {
		filename := r.FormValue("filename")
		keyB64 := r.FormValue("key")
		if filename == "" || keyB64 == "" {
			http.Error(w, "filename and key required", http.StatusBadRequest)
			return
		}
		key, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			http.Error(w, "invalid key", http.StatusBadRequest)
			return
		}
		encPath := filepath.Join("./encrypted", filename)
		data, err := decryptFile(encPath, key)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decrypt: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+strings.TrimSuffix(filename, ".enc"))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data)
	})

	fmt.Printf("Server is running on http://%s/", host)
	http.ListenAndServe(host, r)}
