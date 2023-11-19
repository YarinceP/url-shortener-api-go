package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/yarincep/url-shortener-api-go/api/routes"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	app := fiber.New()

	app.Use(logger.New())
	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveUrlHandler)
	app.Post("/api/v1", routes.ShortenURL)
}
