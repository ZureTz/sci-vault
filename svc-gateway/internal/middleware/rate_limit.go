package middleware

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gateway/pkg/cache"

	"github.com/gin-gonic/gin"
)

//go:embed scripts/rate_limit.lua
var rateLimitScript string

// StrictRateLimit handles rate limiting by IP, Normalized Email, and IP+Email using an atomic Lua script.
func StrictRateLimit(cacheConn *cache.CacheConnector, prefix string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enforce safety: timeout the context
		ctx := c.Request.Context()

		ip := c.ClientIP()

		// Read body with limit (8KB max to mitigate large body attacks)
		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 8192))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Restore the request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			Email string `json:"email"`
		}

		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Email == "" {
			c.Next()
			return
		}

		// Normalize and hash email
		email := strings.ToLower(strings.TrimSpace(req.Email))
		hashBytes := sha256.Sum256([]byte(email))
		emailHash := hex.EncodeToString(hashBytes[:])

		keyIP := fmt.Sprintf("ratelimit:%s:ip:%s", prefix, ip)
		keyEmail := fmt.Sprintf("ratelimit:%s:email:%s", prefix, emailHash)
		keyComposite := fmt.Sprintf("ratelimit:%s:ip_email:%s:%s", prefix, ip, emailHash)

		// Execute Lua script
		keys := []string{keyIP, keyEmail, keyComposite}
		windowMs := window.Milliseconds()

		result, err := cacheConn.Client().Eval(ctx, rateLimitScript, keys, limit, windowMs).Result()
		if err != nil {
			slog.Error("Rate limit redis execution failed", "err", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
			return
		}

		resArr := result.([]interface{})
		count := resArr[0].(int64)
		ttlMs := resArr[1].(int64)
		// blockedKey := resArr[2].(string)

		// Standard rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(max(0, limit-count), 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix()+(ttlMs/1000), 10))

		if count > limit {
			retryAfter := ttlMs / 1000
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
