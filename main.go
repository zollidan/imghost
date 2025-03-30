package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)


func main() {
	file_log, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file_log)

    app := fiber.New(fiber.Config{
		AppName: "ImageHost v0.0.1",
	})

	api := app.Group("/api")
	
	api.Static("/static", "./public")

	file := api.Group("/file")

	file.Get("/list", func (c *fiber.Ctx) error {
		
		ListObjects()

		return c.SendStatus(fiber.StatusOK)
	})

	file.Post("/upload", func(c *fiber.Ctx) error {
	
		file, err := c.FormFile("document")
		if err != nil {
			return err
		}
		
		c.SaveFile(file, fmt.Sprintf("./public/%s", file.Filename))

		return c.Status(fiber.StatusCreated).SendString(fmt.Sprintf("File url: http://localhost:3000/api/static/%s", file.Filename))
	})


    log.Fatal(app.Listen("127.0.0.1:3000"))
}


func ListObjects() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден или не удалось загрузить")
	}	

	cfg, err := config.LoadDefaultConfig(context.TODO(), 
	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		os.Getenv("AWS_ID"), 
		os.Getenv("AWS_SECRET_KEY"), 
		"")),
	config.WithBaseEndpoint(os.Getenv("AWS_ENDPOINT")),
	config.WithRegion(os.Getenv("AWS_REGION")),
	)

	if err != nil {
		log.Fatalf("Ошибка конфигурации AWS: %v", err)
	}

    client := s3.NewFromConfig(cfg)

	resp, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
	})
	
	if err != nil {
		log.Fatalf("Ошибка при запросе к S3: %v", err)
	}

	for _, obj := range resp.Contents {
		fmt.Printf("Файл: %s | Размер: %d байт | Последнее изменение: %s\n",
			aws.ToString(obj.Key),
			obj.Size,
			obj.LastModified.Local().Format("2006-01-02 15:04:05"),
		)
	}
}