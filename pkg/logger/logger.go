package logger

import (
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例（otelzap包装的logger）
	Logger *otelzap.Logger
	// Sugar 全局 Sugar Logger 实例
	Sugar *zap.SugaredLogger
	// baseLogger 基础 zap logger（供需要原始logger的地方使用）
	baseLogger *zap.Logger
)

// Config 日志配置
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     string // stdout, file
	FilePath   string
	MaxSize    int // MB
	MaxBackups int // 保留的旧日志文件数量
	MaxAge     int // 天数
	Compress   bool
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 设置日志级别
	level := parseLevel(cfg.Level)

	// 设置编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" {
		// 文件输出
		writeSyncer = getLogWriter(cfg)
	} else if cfg.Output == "both" {
		// 同时输出到控制台和文件
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			getLogWriter(cfg),
		)
	} else {
		// 默认输出到控制台
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建基础日志实例
	baseLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zapcore.ErrorLevel))

	// 使用 otelzap 包装，自动将日志关联到 OpenTelemetry traces
	Logger = otelzap.New(baseLogger,
		otelzap.WithMinLevel(zapcore.InfoLevel), // 只将 Info 及以上级别的日志发送到 OpenTelemetry
		otelzap.WithCallerDepth(1),              // 正确的调用栈深度
	)
	Sugar = baseLogger.Sugar()

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getLogWriter 获取日志写入器
func getLogWriter(cfg *Config) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// Sync 刷新日志缓冲区
func Sync() {
	if baseLogger != nil {
		_ = baseLogger.Sync()
	}
	if Sugar != nil {
		_ = Sugar.Sync()
	}
}

// Debug 快捷方法
func Debug(msg string, fields ...zap.Field) {
	baseLogger.Debug(msg, fields...)
}

// Info 快捷方法
func Info(msg string, fields ...zap.Field) {
	baseLogger.Info(msg, fields...)
}

// Warn 快捷方法
func Warn(msg string, fields ...zap.Field) {
	baseLogger.Warn(msg, fields...)
}

// Error 快捷方法
func Error(msg string, fields ...zap.Field) {
	baseLogger.Error(msg, fields...)
}

// Fatal 快捷方法
func Fatal(msg string, fields ...zap.Field) {
	baseLogger.Fatal(msg, fields...)
}

// Debugf 格式化快捷方法
func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

// Infof 格式化快捷方法
func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

// Warnf 格式化快捷方法
func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

// Errorf 格式化快捷方法
func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

// Fatalf 格式化快捷方法
func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}
