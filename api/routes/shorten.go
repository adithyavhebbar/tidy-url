package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/adithyavhebbar/tidy-url/database"
	"github.com/adithyavhebbar/tidy-url/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateLimitReset int           `json:"rate_limit_reset"`
	XRateLimiting   int           `json:"rate_limit"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(Request)
	rdb := database.CreateClient(1)
	defer rdb.Close()

	val, err := rdb.Get(database.Ctx, c.IP()).Result()

	if err == redis.Nil {
		fmt.Println("[INFO]: New IP discovered. Creating API QUOTA for IP")
		_ = rdb.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else if err != nil {
		fmt.Println("[ERROR]: Creating API quota for IP failed. Cannot connect to database")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot process the request"})
	} else {
		val, _ = rdb.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit, _ := rdb.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": fmt.Sprint("Limit reached. Please try Later after %d secods", limit/time.Nanosecond/time.Minute),
				"rate limit": limit / time.Nanosecond / time.Minute})
		}
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Cannot parse URL"})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	// Check for custom domain short

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(1)

	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()

	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": fmt.Sprint("Custom short %s is already used", body.CustomShort)})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	result, err := r.Set(database.Ctx, id, body.URL, time.Duration(body.Expiry)*30*60*time.Second).Result()

	fmt.Println(result)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot get shorten URL"})
	}

	resp := Response{
		URL:             id,
		CustomShort:     body.CustomShort,
		Expiry:          body.Expiry,
		XRateLimitReset: 30,
		XRateLimiting:   10,
	}

	val, _ = r.Get(database.Ctx, c.IP()).Result()

	resp.XRateLimiting, _ = strconv.Atoi(val)

	ttl, _ := r.TTL(database.Ctx, c.IP()).Result()

	resp.XRateLimitReset = int(ttl / time.Nanosecond / time.Minute)

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	_ = rdb.Decr(database.Ctx, c.IP()).Err()
	return c.Status(fiber.StatusOK).JSON(resp)
}
