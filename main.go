package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	ratelimiter2 "ratelimiter/internal/ratelimiter"
	"ratelimiter/internal/ratelimiter_middleware"
)

func main() {
	e := echo.New()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// tokenBucketRateLimiter := ratelimiter2.NewTokenBucketRateLimiter(10, 1, redisClient)
	// fixedWindowRateLimiter := ratelimiter2.NewFixedWindowRateLimiter(60, 60*time.Second, redisClient)
	slidingWindowRateLimiter := ratelimiter2.NewSlidingWindowRateLimiter(60, 60*time.Second, redisClient)

	rateLimitMiddleware := ratelimiter_middleware.NewRateLimitMiddleware(slidingWindowRateLimiter)

	e.GET("/sliding_window_v2", func(c echo.Context) error {
		return c.String(http.StatusOK, "Sliding window! Let's Go!")
	}, rateLimitMiddleware.RateLimit())

	e.Start(":8080")
}
