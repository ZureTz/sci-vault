package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"

	// Custom colors for slog levels
	slogInfo  = "\033[32m" // Green
	slogWarn  = "\033[33m" // Yellow
	slogError = "\033[31m" // Red
	slogDebug = "\033[36m" // Cyan
)

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	var levelColor string

	switch r.Level {
	case slog.LevelDebug:
		levelColor = slogDebug
	case slog.LevelInfo:
		levelColor = slogInfo
	case slog.LevelWarn:
		levelColor = slogWarn
	case slog.LevelError:
		levelColor = slogError
	default:
		levelColor = reset
	}

	// Print the timestamp, colored level, and message
	fmt.Printf("%s %s%-5s%s %s",
		r.Time.Format("15:04:05.000"),
		levelColor, level, reset,
		r.Message,
	)

	// Print all attributes in the record
	r.Attrs(func(a slog.Attr) bool {
		fmt.Printf(" %s%s%s=%v", cyan, a.Key, reset, a.Value.Any())
		return true
	})

	fmt.Println()
	return nil
}

// GinLogger returns a gin.LogFormatter func with colored output based on status code.
func GinLogger(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	statusColor = colorForStatus(param.StatusCode)
	methodColor = colorForMethod(param.Method)
	resetColor = reset

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

// Setup initializes the global slog default logger.
//
// level:  "debug" | "info" | "warn" | "error"  (default: "info")
// format: "json"  | "text"                      (default: "json")
func Setup(level, format string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = &customHandler{
			Handler: slog.NewTextHandler(os.Stdout, opts),
		}
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
