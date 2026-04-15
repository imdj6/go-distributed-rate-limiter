package limiter

import (
	"context"
	"fmt"
	"my-gateway/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

// 1. THE SERVICE STRUCT (The Class)
// We hold a pointer to our database, and a pointer to the compiled Lua script.
type RateLimiterService struct {
	store  *config.RedisStore
	script *redis.Script
}

// 2. THE CONSTRUCTOR
// We inject the database store pointer into the service when we create it.
func NewRateLimiterService(store *config.RedisStore) *RateLimiterService {

	// The Lua Script (Token Bucket with Lazy Evaluation)
	luaScript := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refillRate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local requested = 1

		local bucket = redis.call("HMGET", key, "tokens", "last_refill")
		local tokens = tonumber(bucket[1])
		local lastRefill = tonumber(bucket[2])

		if not tokens then
			tokens = capacity
			lastRefill = now
		else
			local timePassed = now - lastRefill
			local refilledTokens = math.floor(timePassed * refillRate)
			tokens = math.min(capacity, tokens + refilledTokens)
			if refilledTokens > 0 then
				lastRefill = now
			end
		end

		if tokens >= requested then
			tokens = tokens - requested
			redis.call("HMSET", key, "tokens", tokens, "last_refill", lastRefill)
			redis.call("EXPIRE", key, 60)
			return {1, tokens} 
		else
			redis.call("HMSET", key, "tokens", tokens, "last_refill", lastRefill)
			redis.call("EXPIRE", key, 60)
			return {0, tokens}
		end
	`

	// We compile the script ONCE during boot-up.
	script := redis.NewScript(luaScript)

	return &RateLimiterService{
		store:  store,
		script: script,
	}
}

// 3. THE BEHAVIOR (The Method)
// Returns: isAllowed (bool), remainingTokens (int64), and an error
func (rl *RateLimiterService) CheckLimit(ipAddress string) (bool, int64, error) {

	redisKey := fmt.Sprintf("rate_limit:%s", ipAddress)
	capacity := 1
	refillRate := 1
	now := time.Now().Unix()

	// A stopwatch for this specific network call (Fail-Fast)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the script in Redis
	result, err := rl.script.Run(ctx, rl.store.Client, []string{redisKey}, capacity, refillRate, now).Result()

	if err != nil {
		// FAIL OPEN PHILOSOPHY:
		// If Redis is dead, we log the error, but we return true (Allowed)
		// so we don't block legitimate users.
		fmt.Printf("⚠️ Redis Error (Failing Open): %v\n", err)
		return true, 0, err
	}

	// TYPE ASSERTION (Go's version of 'as any[]' in TypeScript)
	data := result.([]interface{})
	allowedInt := data[0].(int64)
	remainingTokens := data[1].(int64)

	isAllowed := allowedInt == 1

	return isAllowed, remainingTokens, nil
}
