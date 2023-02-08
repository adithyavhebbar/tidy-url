package main

import (
	"fmt"
	"log"
	"os"

	"github.com/adithyavhebbar/tidy-url/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables
	err := godotenv.Load()

	if err != nil {
		fmt.Println("[ERROR]: Error Getting ENV variables")
	}

	app := fiber.New()

	app.Use(logger.New())

	setUpRoutes(app)
	port := os.Getenv("APP_PORT")
	log.Fatal(app.Listen(port))
}

func setUpRoutes(app *fiber.App) {
	app.Get("/v1/url/", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)

}
