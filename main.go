package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
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
	
}