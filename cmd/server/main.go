package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ricardo/mcp-ocr-server/internal/config"
	"github.com/ricardo/mcp-ocr-server/internal/server"
	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
	version    = flag.Bool("version", false, "Print version information")
)

const (
	appVersion = "1.0.0"
	appName    = "mcp-ocr-server"
)

func main() {
	flag.Parse()

	// 打印版本信息
	if *version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := initLogger(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting MCP OCR Server",
		zap.String("version", appVersion),
		zap.String("config", *configPath),
	)

	// 创建服务器
	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器 (在 goroutine 中)
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	// 等待信号或错误
	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	case err := <-errChan:
		logger.Error("Server error", zap.Error(err))
	}

	// 优雅关闭
	if err := srv.Close(); err != nil {
		logger.Error("Error closing server", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// loadConfig 加载配置文件
func loadConfig(path string) (*config.Config, error) {
	// 如果配置文件不存在，使用默认配置
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Config file not found, using default config\n")
		return config.GetDefault(), nil
	}

	// 加载配置文件
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// initLogger 初始化日志系统
func initLogger(cfg *config.Config) error {
	loggerConfig := logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		OutputPath: cfg.Logger.OutputPath,
	}

	return logger.Init(loggerConfig)
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	// 尝试多个位置
	paths := []string{
		"configs/config.yaml",
		"./config.yaml",
		"/etc/mcp-ocr-server/config.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return "configs/config.yaml"
}