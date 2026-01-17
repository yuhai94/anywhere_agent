package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局zap logger实例
var Logger *zap.Logger

// Init 初始化日志系统
func Init(logDir, logLevel string, maxSize, maxBackups, maxAge int) {
	// 配置日志格式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置日志级别
	level := zap.InfoLevel
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	// 创建日志目录（如果不存在）
	if err := os.MkdirAll(logDir, 0755); err != nil {
		zap.NewNop().Error("Failed to create log directory", zap.Error(err))
	}

	// 配置日志输出到文件
	logFilePath := logDir + "/agent.log"
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSize, // MB
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // 天
		Compress:   true,
	})

	// 配置同时输出到控制台
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 创建zap核心
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, level),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), consoleWriter, level),
	)

	// 创建logger实例
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

// Sync 刷新日志
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// String 创建字符串类型的zap.Field
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Int 创建整数类型的zap.Field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 创建int64类型的zap.Field
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Bool 创建布尔类型的zap.Field
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Time 创建时间类型的zap.Field
func Time(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}

// Error 创建错误类型的zap.Field
func ErrorField(key string, value error) zap.Field {
	return zap.Error(value)
}
