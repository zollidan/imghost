package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileInfo хранит информацию о загруженном файле
type FileInfo struct {
	Path      string
	UploadTime time.Time
}

// FileStorage управляет временным хранением файлов
type FileStorage struct {
	mu        sync.RWMutex
	files     map[string]*FileInfo
	uploadDir string
}

// NewFileStorage создает новое хранилище файлов
func NewFileStorage(uploadDir string) *FileStorage {
	return &FileStorage{
		files:     make(map[string]*FileInfo),
		uploadDir: uploadDir,
	}
}

// StartCleanup запускает горутину для периодической очистки старых файлов
func (fs *FileStorage) StartCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Проверка каждую минуту
		defer ticker.Stop()

		for range ticker.C {
			fs.cleanupOldFiles()
		}
	}()
}

// cleanupOldFiles удаляет файлы старше 15 минут
func (fs *FileStorage) cleanupOldFiles() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now()
	for id, info := range fs.files {
		if now.Sub(info.UploadTime) > 15*time.Minute {
			// Удаляем файл с диска
			if err := os.Remove(info.Path); err != nil {
				log.Printf("Ошибка при удалении файла %s: %v", info.Path, err)
			}
			// Удаляем из карты
			delete(fs.files, id)
			log.Printf("Файл %s удален после 15 минут", id)
		}
	}
}

// SaveFile сохраняет загруженный файл
func (fs *FileStorage) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	// Генерируем уникальный ID для файла
	fileID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
	
	// Открываем загруженный файл
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Создаем путь для сохранения
	filePath := filepath.Join(fs.uploadDir, fileID)
	
	// Создаем файл на диске
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Сохраняем информацию о файле
	fs.mu.Lock()
	fs.files[fileID] = &FileInfo{
		Path:       filePath,
		UploadTime: time.Now(),
	}
	fs.mu.Unlock()

	return fileID, nil
}

// GetFile возвращает путь к файлу по ID
func (fs *FileStorage) GetFile(fileID string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	info, exists := fs.files[fileID]
	if !exists {
		return "", false
	}
	return info.Path, true
}

var storage *FileStorage

// uploadHandler обрабатывает загрузку файлов
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart форму (максимум 10MB)
	r.ParseMultipartForm(10 << 20)

	// Получаем файл из формы
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Сохраняем файл
	fileID, err := storage.SaveFile(handler)
	if err != nil {
		http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
		return
	}

	// Возвращаем ID файла
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"file_id": "%s", "message": "Файл будет удален через 15 минут"}`, fileID)
}

// downloadHandler обрабатывает скачивание файлов
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Не указан ID файла", http.StatusBadRequest)
		return
	}

	filePath, exists := storage.GetFile(fileID)
	if !exists {
		http.Error(w, "Файл не найден или уже удален", http.StatusNotFound)
		return
	}

	// Отправляем файл
	http.ServeFile(w, r, filePath)
}

// indexHandler показывает простую HTML форму для загрузки
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html, err := os.Open("templates/index.html")
    if err != nil{
        return 
    }
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func main() {
	// Создаем директорию для загрузок
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Не удалось создать директорию для загрузок:", err)
	}

	// Инициализируем хранилище
	storage = NewFileStorage(uploadDir)
	storage.StartCleanup()

	// Настраиваем обработчики
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)

	// Запускаем сервер
	port := ":8080"
	log.Printf("Сервер запущен на http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}