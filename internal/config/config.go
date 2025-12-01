package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	OCR           OCRConfig           `yaml:"ocr"`
	Preprocessing PreprocessingConfig `yaml:"preprocessing"`
	Performance   PerformanceConfig   `yaml:"performance"`
	Logger        LoggerConfig        `yaml:"logger"`
}

// ServerConfig MCP Server 配置
type ServerConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// OCRConfig OCR 引擎配置
type OCRConfig struct {
	Engine          string   `yaml:"engine"`           // tesseract
	Language        string   `yaml:"language"`         // eng+chi_sim+chi_tra+jpn
	DataPath        string   `yaml:"data_path"`        // tessdata 路径
	PageSegMode     int      `yaml:"page_seg_mode"`    // 页面分割模式 (3=全自动)
	EngineMode      int      `yaml:"engine_mode"`      // 引擎模式 (3=默认)
	Whitelist       string   `yaml:"whitelist"`        // 字符白名单
	SupportedLangs  []string `yaml:"supported_langs"`  // 支持的语言列表
	MaxImageSize    int64    `yaml:"max_image_size"`   // 最大图像大小(字节)
	Timeout         int      `yaml:"timeout"`          // OCR 超时时间(秒)
}

// PreprocessingConfig 图像预处理配置
type PreprocessingConfig struct {
	Enabled           bool    `yaml:"enabled"`             // 是否启用预处理
	AutoMode          bool    `yaml:"auto_mode"`           // 自动分析模式
	Grayscale         bool    `yaml:"grayscale"`           // 灰度化
	Denoise           bool    `yaml:"denoise"`             // 降噪
	DenoiseStrength   int     `yaml:"denoise_strength"`    // 降噪强度 (3-11)
	Binarization      bool    `yaml:"binarization"`        // 二值化
	BinarizationMode  string  `yaml:"binarization_mode"`   // 二值化模式: otsu, adaptive
	AdaptiveBlockSize int     `yaml:"adaptive_block_size"` // 自适应二值化块大小
	AdaptiveC         float64 `yaml:"adaptive_c"`          // 自适应二值化常数
	DeskewCorrection  bool    `yaml:"deskew_correction"`   // 倾斜校正
	DeskewAngleLimit  float64 `yaml:"deskew_angle_limit"`  // 倾斜角度限制
	Resize            bool    `yaml:"resize"`              // 是否调整大小
	ResizeWidth       int     `yaml:"resize_width"`        // 调整后的宽度
	ResizeHeight      int     `yaml:"resize_height"`       // 调整后的高度
	QualityThresholds struct {
		Sharpness  float64 `yaml:"sharpness"`  // 清晰度阈值
		Contrast   float64 `yaml:"contrast"`   // 对比度阈值
		Brightness float64 `yaml:"brightness"` // 亮度阈值 (最小值)
	} `yaml:"quality_thresholds"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	WorkerPoolSize  int  `yaml:"worker_pool_size"`  // Worker 池大小
	QueueSize       int  `yaml:"queue_size"`        // 任务队列大小
	CacheEnabled    bool `yaml:"cache_enabled"`     // 是否启用缓存
	CacheSize       int  `yaml:"cache_size"`        // 缓存大小
	CacheTTL        int  `yaml:"cache_ttl"`         // 缓存 TTL (秒)
	ResourcePooling bool `yaml:"resource_pooling"`  // 是否启用资源池
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `yaml:"level"`       // 日志级别
	Format     string `yaml:"format"`      // 输出格式
	OutputPath string `yaml:"output_path"` // 输出路径
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 处理路径
	if err := cfg.ProcessPaths(); err != nil {
		return nil, fmt.Errorf("failed to process paths: %w", err)
	}

	return &cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证 OCR 配置
	if c.OCR.Engine != "tesseract" {
		return fmt.Errorf("unsupported OCR engine: %s", c.OCR.Engine)
	}

	if c.OCR.MaxImageSize <= 0 {
		return fmt.Errorf("invalid max_image_size: %d", c.OCR.MaxImageSize)
	}

	if c.OCR.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: %d", c.OCR.Timeout)
	}

	// 验证性能配置
	if c.Performance.WorkerPoolSize <= 0 {
		return fmt.Errorf("invalid worker_pool_size: %d", c.Performance.WorkerPoolSize)
	}

	if c.Performance.QueueSize <= 0 {
		return fmt.Errorf("invalid queue_size: %d", c.Performance.QueueSize)
	}

	if c.Performance.CacheEnabled && c.Performance.CacheSize <= 0 {
		return fmt.Errorf("invalid cache_size: %d", c.Performance.CacheSize)
	}

	// 验证日志配置
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logger.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[c.Logger.Format] {
		return fmt.Errorf("invalid log format: %s", c.Logger.Format)
	}

	return nil
}

// ProcessPaths 处理配置中的路径
func (c *Config) ProcessPaths() error {
	// 处理 tessdata 路径
	if c.OCR.DataPath != "" {
		absPath, err := filepath.Abs(c.OCR.DataPath)
		if err != nil {
			return fmt.Errorf("invalid data_path: %w", err)
		}
		c.OCR.DataPath = absPath
	}

	// 处理日志输出路径
	if c.Logger.OutputPath != "" && c.Logger.OutputPath != "stdout" && c.Logger.OutputPath != "stderr" {
		absPath, err := filepath.Abs(c.Logger.OutputPath)
		if err != nil {
			return fmt.Errorf("invalid log output_path: %w", err)
		}
		c.Logger.OutputPath = absPath

		// 确保日志目录存在
		logDir := filepath.Dir(absPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	return nil
}

// GetDefault 获取默认配置
func GetDefault() *Config {
	return &Config{
		Server: ServerConfig{
			Name:        "mcp-ocr-server",
			Version:     "1.0.0",
			Description: "Production-grade OCR MCP Server with intelligent preprocessing",
		},
		OCR: OCRConfig{
			Engine:         "tesseract",
			Language:       "eng+chi_sim+chi_tra+jpn",
			DataPath:       "/usr/local/share/tessdata",
			PageSegMode:    3,
			EngineMode:     3,
			Whitelist:      "",
			SupportedLangs: []string{"eng", "chi_sim", "chi_tra", "jpn"},
			MaxImageSize:   10 * 1024 * 1024, // 10MB
			Timeout:        30,
		},
		Preprocessing: PreprocessingConfig{
			Enabled:           true,
			AutoMode:          true,
			Grayscale:         true,
			Denoise:           true,
			DenoiseStrength:   5,
			Binarization:      true,
			BinarizationMode:  "otsu",
			AdaptiveBlockSize: 11,
			AdaptiveC:         2.0,
			DeskewCorrection:  true,
			DeskewAngleLimit:  10.0,
			Resize:            false,
			ResizeWidth:       0,
			ResizeHeight:      0,
		},
		Performance: PerformanceConfig{
			WorkerPoolSize:  4,
			QueueSize:       100,
			CacheEnabled:    true,
			CacheSize:       100,
			CacheTTL:        3600,
			ResourcePooling: true,
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "console",
			OutputPath: "stdout",
		},
	}
}