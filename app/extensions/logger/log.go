package logger

import (
	"context"
	"runtime/debug"

	"github.com/daheige/gomux/app/helper"

	"github.com/daheige/thinkgo/gutils"
	"github.com/daheige/thinkgo/logger"
)

/**
{
    "level":"info",
    "time_local":"2019-11-24T18:18:49.978+0800",
    "msg":"exec end",
    "plat":"web",
    "request_method":"GET",
    "trace_line":49,
    "tag":"api_v1_hello",
    "ip":"127.0.0.1:37820",
    "ua":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.80 Safari/537.36",
    "trace_file":"/web/go/hg-mux/app/middleware/RequestWare.go",
    "request_uri":"/api/v1/hello",
    "log_id":"064dee71589e70fa88e8b296128dad95",
    "options":{
        "exec_time":0.001184617
    }
}
*/

func writeLog(ctx context.Context, levelName string, message string, options map[string]interface{}) {
	reqUri := getStringByCtx(ctx, "request_uri")
	logId := getStringByCtx(ctx, "log_id")
	if logId == "" {
		logId = gutils.RndUuid()
	}

	ua := getStringByCtx(ctx, "user_agent")
	logInfo := map[string]interface{}{
		"request_uri":    reqUri,
		"log_id":         logId,
		"options":        options,
		"ip":             getStringByCtx(ctx, "client_ip"),
		"ua":             ua,
		"plat":           helper.GetDeviceByUa(ua), // 当前设备匹配
		"request_method": getStringByCtx(ctx, "request_method"),
	}

	switch levelName {
	case "info":
		logger.Info(message, logInfo)
	case "debug":
		logger.Debug(message, logInfo)
	case "warn":
		logger.Warn(message, logInfo)
	case "error":
		logger.Error(message, logInfo)
	case "emergency":
		logger.DPanic(message, logInfo)
	case "fatal":
		logger.Fatal(message, logInfo)
	default:
		logger.Info(message, logInfo)
	}
}

func getStringByCtx(ctx context.Context, key string) string {
	return helper.GetStringByCtx(ctx, key)
}

// Info info log.
func Info(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "info", message, context)
}

// Debug debug log.
func Debug(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "debug", message, context)
}

// Warn warn log.
func Warn(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "warn", message, context)
}

// Error error log.
func Error(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "error", message, context)
}

// Emergency 致命错误或panic捕获，程序继续运行不会崩溃
func Emergency(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "emergency", message, context)
}

// Fatal 抛出致命操作，程序退出
func Fatal(ctx context.Context, message string, context map[string]interface{}) {
	writeLog(ctx, "fatal", message, context)
}

// Recover 异常捕获处理
func Recover(ctx context.Context) {
	if err := recover(); err != nil {
		Emergency(ctx, "exec panic", map[string]interface{}{
			"error":       err,
			"error_trace": string(debug.Stack()),
		})
	}
}
