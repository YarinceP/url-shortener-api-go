package routes

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yarincep/url-shortener-api-go/api/database"
	"github.com/yarincep/url-shortener-api-go/api/helpers"
	"os"
	"strconv"
	"time"
)

func ShortenURL(c *fiber.Ctx) error {
	request := new(UrlShortenerRequest)

	err := c.BodyParser(&request)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	//Implement rate limiting
	r := database.CreateClient(1)
	defer r.Close()
	result, err := r.Get(database.DBContext, c.IP()).Result()
	if err == nil {
		errorOnSetValue := r.Set(database.DBContext, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
		if errorOnSetValue != nil {
			fmt.Println(errorOnSetValue, ": Error on set value ip quota on redis")
		}
	}
	if err != nil {
		result, _ = r.Get(database.DBContext, c.IP()).Result()
		counter, _ := strconv.Atoi(result)

		if counter <= 0 {
			limit, _ := r.TTL(database.DBContext, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rate limit exceeded", "rate_limit_rest": limit / time.Nanosecond / time.Minute})
		}

		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot connect to DB"})
	}

	//Check if the input is an actual url
	if !helpers.IsUrlValid(request.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}

	//Check for domain error
	if !helpers.RemoveDomainError(request.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "domain error"})
	}

	//enforce https, SSL
	request.URL = helpers.EnforceHttp(request.URL)

	//validate custom short url field
	var id string

	if request.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = request.CustomShort
	}

	r2 := database.CreateClient(0)
	defer r2.Close()

	result, _ = r2.Get(database.DBContext, id).Result()
	if result != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Url custom short is already in use"})
	}

	if request.Expiry == 0 {
		request.Expiry = 24
	}

	err = r2.Set(database.DBContext, id, request.URL, request.Expiry).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to server"})
	}

	response := UrlShortenerResponse{
		URL:           request.URL,
		CustomShort:   "",
		Expiry:        request.Expiry,
		RateRemaining: 10,
		RateLimitRest: 30,
	}

	//Decrement time
	r.Decr(database.DBContext, c.IP())

	result, _ = r.Get(database.DBContext, c.IP()).Result()
	response.RateRemaining, _ = strconv.Atoi(result)

	ttl, _ := r.TTL(database.DBContext, c.IP()).Result()
	response.RateLimitRest = ttl / time.Nanosecond / time.Minute

	response.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(response)
}

func ResolveUrlHandler(c *fiber.Ctx) error {

	url := c.Params("url")

	redisClient := database.CreateClient(0)

	defer func(redisClient *redis.Client) {
		err := redisClient.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(redisClient)

	result, err := redisClient.Get(database.DBContext, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "short url not founded in the redis database"})
	}
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot connect to DB"})
	}

	redisClientIncrement := database.CreateClient(1)

	defer func(redisClient *redis.Client) {
		err := redisClient.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(redisClientIncrement)

	_ = redisClientIncrement.Incr(database.DBContext, "counter")

	return c.Redirect(result, fiber.StatusMovedPermanently)
}
