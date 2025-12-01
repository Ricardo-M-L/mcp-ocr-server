package ocr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/ricardo/mcp-ocr-server/pkg/errors"
)

// TesseractCGoEngine Tesseract CGo 引擎实现
type TesseractCGoEngine struct {
	client *gosseract.Client
	config EngineConfig
	mu     sync.Mutex
}

// NewTesseractCGoEngine 创建新的 Tesseract CGo 引擎
func NewTesseractCGoEngine() *TesseractCGoEngine {
	return &TesseractCGoEngine{
		client: gosseract.NewClient(),
	}
}

// Initialize 初始化引擎
func (e *TesseractCGoEngine) Initialize(config EngineConfig) error {
	e.config = config

	// 设置 tessdata 路径
	if config.TessdataPath != "" {
		e.client.SetTessdataPrefix(config.TessdataPath)
	}

	// 设置默认语言
	if len(config.Languages) > 0 {
		e.client.SetLanguage(config.Languages[0])
	}

	// 设置页面分割模式
	if config.DefaultPSM > 0 {
		e.client.SetPageSegMode(gosseract.PSM(config.DefaultPSM))
	}

	return nil
}

// ExtractText 从图像文件提取文字
func (e *TesseractCGoEngine) ExtractText(ctx context.Context, imagePath string, opts Options) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	start := time.Now()

	// 检查上下文取消
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), errors.ErrTimeout, "context cancelled")
	default:
	}

	// 设置语言
	if opts.Language != "" {
		e.client.SetLanguage(opts.Language)
	}

	// 设置 PSM
	if opts.PSM > 0 {
		e.client.SetPageSegMode(gosseract.PSM(opts.PSM))
	}

	// 设置图像
	if err := e.client.SetImage(imagePath); err != nil {
		return nil, errors.Wrap(err, errors.ErrOCREngineFailed, "failed to set image")
	}

	// 提取文本
	text, err := e.client.Text()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrOCREngineFailed, "failed to extract text")
	}

	// 获取置信度
	conf, err := e.client.MeanConfidence()
	if err != nil {
		conf = 0
	}

	duration := time.Since(start)

	return &Result{
		Text:       text,
		Confidence: float64(conf),
		Language:   opts.Language,
		Duration:   duration,
		Words:      []WordDetail{},
	}, nil
}

// ExtractTextFromBytes 从字节数组提取文字
func (e *TesseractCGoEngine) ExtractTextFromBytes(ctx context.Context, imageData []byte, opts Options) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	start := time.Now()

	// 检查上下文取消
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), errors.ErrTimeout, "context cancelled")
	default:
	}

	// 设置语言
	if opts.Language != "" {
		e.client.SetLanguage(opts.Language)
	}

	// 设置 PSM
	if opts.PSM > 0 {
		e.client.SetPageSegMode(gosseract.PSM(opts.PSM))
	}

	// 设置图像数据
	if err := e.client.SetImageFromBytes(imageData); err != nil {
		return nil, errors.Wrap(err, errors.ErrOCREngineFailed, "failed to set image from bytes")
	}

	// 提取文本
	text, err := e.client.Text()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrOCREngineFailed, "failed to extract text")
	}

	// 获取置信度
	conf, err := e.client.MeanConfidence()
	if err != nil {
		conf = 0
	}

	duration := time.Since(start)

	return &Result{
		Text:       text,
		Confidence: float64(conf),
		Language:   opts.Language,
		Duration:   duration,
		Words:      []WordDetail{},
	}, nil
}

// Type 获取引擎类型
func (e *TesseractCGoEngine) Type() EngineType {
	return EngineTypeCGo
}

// Close 关闭引擎
func (e *TesseractCGoEngine) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// String 返回引擎描述
func (e *TesseractCGoEngine) String() string {
	return fmt.Sprintf("TesseractCGo(languages=%v)", e.config.Languages)
}