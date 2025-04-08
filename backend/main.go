// main.go
package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Go backend!",
		})
	})

	app.Listen("127.0.0.1:3000")
}
