package ocr

import (
	"context"
	"time"
)

// Engine OCR 引擎接口
type Engine interface {
	// Init 初始化引擎
	Init(config EngineConfig) error

	// RecognizeText 识别图像中的文本
	RecognizeText(ctx context.Context, imageData []byte, opts RecognizeOptions) (*RecognizeResult, error)

	// Close 关闭引擎并释放资源
	Close() error

	// GetSupportedLanguages 获取支持的语言列表
	GetSupportedLanguages() []string
}

// EngineConfig 引擎配置
type EngineConfig struct {
	Language     string        // 语言设置
	DataPath     string        // tessdata 路径
	PageSegMode  int           // 页面分割模式
	EngineMode   int           // 引擎模式
	Whitelist    string        // 字符白名单
	Timeout      time.Duration // 超时时间
}

// RecognizeOptions 识别选项
type RecognizeOptions struct {
	Language    string            // 识别语言 (可覆盖默认配置)
	PageSegMode *int              // 页面分割模式 (可覆盖默认配置)
	Preprocess  bool              // 是否预处理
	Metadata    map[string]string // 额外元数据
}

// RecognizeResult 识别结果
type RecognizeResult struct {
	Text       string            // 识别的文本
	Confidence float64           // 置信度 (0-100)
	Language   string            // 使用的语言
	Duration   time.Duration     // 识别耗时
	Metadata   map[string]string // 额外元数据
}

// BoundingBox 文本边界框
type BoundingBox struct {
	X      int     // X 坐标
	Y      int     // Y 坐标
	Width  int     // 宽度
	Height int     // 高度
	Text   string  // 文本内容
	Conf   float64 // 置信度
}

// DetailedResult 详细识别结果(包含边界框)
type DetailedResult struct {
	Text        string        // 全部文本
	Confidence  float64       // 总体置信度
	BoundingBox []BoundingBox // 文本块边界框
	Duration    time.Duration // 识别耗时
}