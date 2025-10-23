package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	fixedWindowKeyTemplate = "%s:%d" // key:minute_number

	fixedWindowScript = `
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
    		redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return current
	`
)

type FixedWindowRateLimiter struct {
	windowSize  time.Duration
	limit       int
	redis       *redis.Client
	redisScript *redis.Script
}

func NewFixedWindowRateLimiter(
	limit int,
	windowSize time.Duration,
	redisClient *redis.Client,
) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		windowSize:  windowSize,
		limit:       limit,
		redis:       redisClient,
		redisScript: redis.NewScript(fixedWindowScript),
	}
}

func (rl *FixedWindowRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	windowNumber := time.Now().Unix() / int64(rl.windowSize.Seconds())
	res, err := rl.redisScript.Run(
		ctx,
		rl.redis,
		[]string{fmt.Sprintf(fixedWindowKeyTemplate, key, windowNumber)},
		[]any{rl.windowSize.Seconds()},
	).Int()

	fmt.Println(res, " - ", fmt.Sprintf(fixedWindowKeyTemplate, key, windowNumber), time.Now().Unix())

	if err != nil {
		return false, fmt.Errorf("run script: %v", err)
	}

	return res <= rl.limit, nil
}
