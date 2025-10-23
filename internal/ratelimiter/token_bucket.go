package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	tokenBucketScript = `
			local key = KEYS[1]
			local capacity = tonumber(ARGV[1])
			local refill_rate = tonumber(ARGV[2])
			local now = tonumber(ARGV[3])
	
			local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
			local tokens = tonumber(bucket[1]) or capacity
			local last_refill = tonumber(bucket[2]) or now
	
			local time_passed = now - last_refill
			local tokens_refill = math.floor(time_passed * refill_rate / 1e9)
			tokens = math.min(tokens + tokens_refill, capacity)
	
			if tokens >= 1 then
				tokens = tokens - 1
				redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
				redis.call('EXPIRE', key, math.ceil(capacity / refill_rate))
				return 1
			else
				return 0
			end`
)

type TokenBucketRateLimiter struct {
	capacity    int
	fillRate    float64
	redis       *redis.Client
	redisScript *redis.Script
}

func NewTokenBucketRateLimiter(
	capacity int,
	fillRate float64,
	redisClient *redis.Client,
) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		capacity:    capacity,
		fillRate:    fillRate,
		redis:       redisClient,
		redisScript: redis.NewScript(tokenBucketScript),
	}
}

func (rl *TokenBucketRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	res, err := rl.redisScript.Run(
		ctx,
		rl.redis,
		[]string{key},
		[]any{rl.capacity, rl.fillRate, time.Now().Unix()},
	).Int()

	if err != nil {
		return false, fmt.Errorf("run script: %v", err)
	}

	if res == 0 {
		return false, nil
	}

	return true, nil
}
