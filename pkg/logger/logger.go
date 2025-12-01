package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// Sugar 全局 SugaredLogger 实例
	Sugar *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
	Format     string `yaml:"format"`      // 输出格式: json, console
	OutputPath string `yaml:"output_path"` // 输出路径: stdout, stderr, 或文件路径
}

// Init 初始化日志系统
func Init(cfg Config) error {
	// 解析日志级别
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 选择编码格式
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	writer, err := getWriter(cfg.OutputPath)
	if err != nil {
		return err
	}

	// 创建 Core
	core := zapcore.NewCore(encoder, writer, level)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// getWriter 获取输出 Writer
func getWriter(path string) (zapcore.WriteSyncer, error) {
	if path == "" || path == "stdout" {
		return zapcore.AddSync(os.Stdout), nil
	}
	if path == "stderr" {
		return zapcore.AddSync(os.Stderr), nil
	}

	// 文件输出
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}

// Sync 刷新日志缓冲区
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// Debug 记录 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info 记录 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn 记录 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error 记录 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal 记录 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// With 创建带有额外字段的 Logger
func With(fields ...zap.Field) *zap.Logger {
	if Logger != nil {
		return Logger.With(fields...)
	}
	return zap.NewNop()
}