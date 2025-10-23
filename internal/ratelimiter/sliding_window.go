package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const slidingWindowScript = `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		redis.call('ZREMRANGEBYSCORE', key, 0, window_start-1)
		local count = redis.call('ZCARD', key)

		if count < limit then
			redis.call('ZADD', key, now, now)
			if count == 0 then
       			 redis.call('EXPIRE', key, ARGV[4])
   	 		end
			return 1
		else
			return 0
		end
	`

type SlidingWindowRateLimiter struct {
	limit       int
	windowSize  time.Duration
	redis       *redis.Client
	redisScript *redis.Script
}

func NewSlidingWindowRateLimiter(
	limit int,
	windowSize time.Duration,
	redisClient *redis.Client,
) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		windowSize:  windowSize,
		limit:       limit,
		redis:       redisClient,
		redisScript: redis.NewScript(slidingWindowScript),
	}
}

func (rl *SlidingWindowRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - rl.windowSize.Milliseconds()

	res, err := rl.redisScript.Run(
		ctx,
		rl.redis,
		[]string{key},
		[]any{rl.limit, windowStart, now, rl.windowSize.Seconds()},
	).Int()
	if err != nil {

	}

	return res == 1, nil
}
