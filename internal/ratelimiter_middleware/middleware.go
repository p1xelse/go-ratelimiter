package ratelimiter_middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ratelimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
type RateLimitMiddleware struct {
	ratelimiter ratelimiter
}

func NewRateLimitMiddleware(ratelimiter ratelimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		ratelimiter: ratelimiter,
	}
}

func (middleware *RateLimitMiddleware) RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			allow, err := middleware.ratelimiter.Allow(c.Request().Context(), c.Request().URL.Path)
			if err != nil {
				allow = true
				log.Println(err)
			}

			if !allow {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
			}

			return next(c)
		}
	}
}
