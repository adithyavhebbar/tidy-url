package routes

import (
	"fmt"

	"github.com/adithyavhebbar/tidy-url/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	que := c.Context().QueryArgs()

	url := string(que.Peek("url"))

	r := database.CreateClient(1)

	defer r.Close()

	longUrl, err := r.Get(database.Ctx, url).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Short URL not found"})
	} else if err != nil {
		fmt.Println("Cannot connect to Database", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Something went wrong"})
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	return c.Redirect(longUrl, 301)
}

func TestServer(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"health": "OK"})
}
