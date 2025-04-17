package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	bucket string
	minioClient *minio.Client
	IS_PROD bool = false //os.Getenv("ENV") == "development" //production
)

type File struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name    string `gorm:"uniqueIndex" json:"name"`
	FileURL string `gorm:"uniqueIndex" json:"file_url"`
}

func initMinIO() {

	if !IS_PROD {
		if err := godotenv.Load("../.env"); err != nil {
			logrus.Info("No .env file found for init MinIO")
		}
	}

	bucket = os.Getenv("MINIO_BUCKET")

	var err error

	minioHost := "localhost"
	if IS_PROD {
		minioHost = "minio"
	}
	
	endpoint := fmt.Sprintf("%s:9000", minioHost)
	
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ID"), os.Getenv("MINIO_SECRET"), ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Проверка/создание бакета
	exists, errBucket := minioClient.BucketExists(context.Background(), bucket)
	if errBucket != nil {
		log.Fatalf("Bucket check failed: %v", errBucket)
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Could not create bucket: %v", err)
		}
	}
}

func InitDatabase() {

	if !IS_PROD {
		if err := godotenv.Load("../.env"); err != nil {
			logrus.Info("No .env file found for init MinIO")
		}
	}

	host := os.Getenv("POSTGRES_HOST")

	if IS_PROD {
		host = "postgres"
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", host, user, password, dbname)
	var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

	db.AutoMigrate(&File{})
}

// get all s3 objects
func GetAllS3Files(c *fiber.Ctx) error {
	ctx := context.Background()
	objects := []map[string]interface{}{}

	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": object.Err.Error(),
			})
		}

		objects = append(objects, fiber.Map{
			"key":          object.Key,
			"size":         object.Size,
			"lastModified": object.LastModified,
		})
	}

	return c.JSON(objects)
}

func GetFileByID(c *fiber.Ctx) error {
	fileID := c.Params("file_id")
	ctx := context.Background()

	object, err := minioClient.GetObject(ctx, bucket, fileID, minio.GetObjectOptions{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Получаем информацию о типе контента
	info, err := object.Stat()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

    c.Set("Content-Type", info.ContentType)
	return c.SendStream(object)
}

// Create
func CreateFile(c *fiber.Ctx) error {
	// Получаем файл из формы
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File not found in form"})
	}

	// Открываем файл
	fileReader, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open uploaded file"})
	}
	defer fileReader.Close()

	// Генерируем UUID
	id := uuid.New()

	// Генерируем имя файла (UUID + расширение)
	ext := filepath.Ext(fileHeader.Filename)
	fileName := id.String() + ext

	// Загружаем в S3
	ctx := context.Background()
	_, err = minioClient.PutObject(ctx, bucket, fileName, fileReader, fileHeader.Size, minio.PutObjectOptions{
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Собираем S3 URL (или ты можешь использовать presigned URL)
	fileURL := fmt.Sprintf("/api/s3/files/%s", fileName)

	// Создаём запись в БД
	file := File{
		ID:      id,
		Name:    fileName,
		FileURL: fileURL,
	}

	if err := db.Create(&file).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(file)
}

// Get all
func GetFiles(c *fiber.Ctx) error {
	var files []File
	db.Find(&files)
	return c.JSON(files)
}

// Get by ID
func GetFile(c *fiber.Ctx) error {
	id := c.Params("id")
	var file File

	if err := db.First(&file, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	return c.JSON(file)
}

// Update
func UpdateFile(c *fiber.Ctx) error {
	id := c.Params("id")
	var file File

	if err := db.First(&file, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	if err := c.BodyParser(&file); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	db.Save(&file)
	return c.JSON(file)
}

// Delete
func DeleteFile(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&File{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

func main() {
    InitDatabase()
    initMinIO()

	app := fiber.New(fiber.Config{
        AppName: "aafbet API",
    })

	app.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, this is aafbet API!",
			"version": "0.0.1",
		})
	})

    // Мидделваре
	// app.Use(jwtware.New(jwtware.Config{
	// 	SigningKey: jwtware.SigningKey{Key: []byte("secret")},
	// }))

    files := app.Group("/api/files")
	files.Post("/", CreateFile)
	files.Get("/", GetFiles)
	files.Get("/:id", GetFile)
	files.Put("/:id", UpdateFile)
	files.Delete("/:id", DeleteFile)

    s3 := app.Group("/api/s3")
    s3.Get("/files", GetAllS3Files)
    s3.Get("/files/:file_id", GetFileByID)

	if !IS_PROD {
		log.Fatal(app.Listen("localhost:3000"))
	}else {
		log.Fatal(app.Listen("0.0.0.0:3000"))
	}

	
}
