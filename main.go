package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)


func main() {
    app := fiber.New(fiber.Config{
		AppName: "ImageHost v0.0.1",
	})

	api := app.Group("/api")
	
	api.Static("/static", "./public")

	api.Post("/file/upload", func(c *fiber.Ctx) error {
		
		// Get first file from form field "document":
		file, err := c.FormFile("document")
		if err != nil {
			return err
		}
		
		c.SaveFile(file, fmt.Sprintf("./public/%s", file.Filename))

		return c.Status(fiber.StatusCreated).SendString(fmt.Sprintf("File url: http://localhost:3000/api/static/%s", file.Filename))
	})


    log.Fatal(app.Listen("localhost:3000"))
}