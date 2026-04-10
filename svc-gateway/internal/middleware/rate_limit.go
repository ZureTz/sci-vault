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

// StrictRateLimit handles rate limiting by IP, Normalized Email, and IP+Email
func StrictRateLimit(cacheConn *cache.CacheConnector, prefix string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()

		// Read and parse request body
		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 8192))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request body too large"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			Email string `json:"email"`
		}

		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Email == "" {
			c.Next()
			return
		}

		// Normalize email
		email := strings.ToLower(strings.TrimSpace(req.Email))
		emailHash := hashEmail(email)

		// Three-layer rate limit keys
		keyIP := fmt.Sprintf("ratelimit:%s:ip:%s", prefix, ip)
		keyEmail := fmt.Sprintf("ratelimit:%s:email:%s", prefix, emailHash)
		keyComposite := fmt.Sprintf("ratelimit:%s:composite:%s:%s", prefix, ip, emailHash)

		keys := []string{keyIP, keyEmail, keyComposite}
		windowMs := window.Milliseconds()

		// Execute Lua script
		result, err := cacheConn.Client().Eval(ctx, rateLimitScript, keys, limit, windowMs).Result()
		if err != nil {
			slog.Error("rate limit script failed", "email", email, "ip", ip, "error", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
			return
		}

		// Safe type conversion from Lua result
		resArray, ok := result.([]interface{})
		if !ok || len(resArray) < 3 {
			slog.Error("unexpected rate limit script result", "result", result, "type", fmt.Sprintf("%T", result))
			c.Next()
			return
		}

		count := toInt64(resArray[0])
		ttlMs := toInt64(resArray[1])
		status := toString(resArray[2])

		slog.Debug("rate limit check", "email", email, "ip", ip, "count", count, "limit", limit, "status", status)

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(max(0, limit-count), 10))
		if ttlMs > 0 {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix()+(ttlMs/1000), 10))
		}

		// Check if rate limit exceeded
		if count > limit {
			retryAfter := max(1, ttlMs/1000)
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			slog.Warn("rate limit exceeded", "email", email, "ip", ip, "count", count, "limit", limit)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}

// hashEmail generates SHA256 hash of email for safe Redis key
func hashEmail(email string) string {
	hash := sha256.Sum256([]byte(email))
	return hex.EncodeToString(hash[:])
}

// toInt64 safely converts various numeric types to int64
func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// toString safely converts interface to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// max returns the larger of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
