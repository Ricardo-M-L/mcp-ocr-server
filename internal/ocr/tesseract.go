package ocr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/otiai10/gosseract/v2"
	ocrErrors "github.com/ricardo/mcp-ocr-server/pkg/errors"
	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// TesseractEngine Tesseract OCR 引擎实现
type TesseractEngine struct {
	config            EngineConfig
	supportedLanguages []string
	clientPool        *sync.Pool
	mu                sync.RWMutex
}

// NewTesseractEngine 创建 Tesseract 引擎实例
func NewTesseractEngine() *TesseractEngine {
	return &TesseractEngine{
		supportedLanguages: []string{"eng", "chi_sim", "chi_tra", "jpn"},
		clientPool: &sync.Pool{
			New: func() interface{} {
				return gosseract.NewClient()
			},
		},
	}
}

// Init 初始化引擎
func (e *TesseractEngine) Init(config EngineConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.config = config

	// 测试 Tesseract 是否可用
	client := e.clientPool.Get().(*gosseract.Client)
	defer e.clientPool.Put(client)

	// 设置配置
	if config.DataPath != "" {
		if err := client.SetTessdataPrefix(config.DataPath); err != nil {
			return ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set tessdata path")
		}
	}

	if err := client.SetLanguage(config.Language); err != nil {
		return ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set language")
	}

	if config.PageSegMode > 0 {
		if err := client.SetPageSegMode(gosseract.PageSegMode(config.PageSegMode)); err != nil {
			return ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set page seg mode")
		}
	}

	if config.Whitelist != "" {
		client.SetWhitelist(config.Whitelist)
	}

	logger.Info("Tesseract engine initialized",
		zap.String("language", config.Language),
		zap.String("data_path", config.DataPath),
		zap.Int("page_seg_mode", config.PageSegMode),
	)

	return nil
}

// RecognizeText 识别图像中的文本
func (e *TesseractEngine) RecognizeText(ctx context.Context, imageData []byte, opts RecognizeOptions) (*RecognizeResult, error) {
	startTime := time.Now()

	// 从池中获取客户端
	client := e.clientPool.Get().(*gosseract.Client)
	defer e.clientPool.Put(client)

	// 应用配置
	if err := e.configureClient(client, opts); err != nil {
		return nil, err
	}

	// 设置图像数据
	if err := client.SetImageFromBytes(imageData); err != nil {
		return nil, ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set image data")
	}

	// 执行 OCR (带超时控制)
	resultChan := make(chan struct {
		text string
		err  error
	}, 1)

	go func() {
		text, err := client.Text()
		resultChan <- struct {
			text string
			err  error
		}{text, err}
	}()

	// 等待结果或超时
	select {
	case <-ctx.Done():
		return nil, ocrErrors.New(ocrErrors.ErrTimeout, "OCR operation timeout")
	case result := <-resultChan:
		if result.err != nil {
			return nil, ocrErrors.Wrap(result.err, ocrErrors.ErrOCREngineFailed, "OCR recognition failed")
		}

		// 获取置信度
		confidence := e.getConfidence(client)

		duration := time.Since(startTime)

		// 构建结果
		ocrResult := &RecognizeResult{
			Text:       result.text,
			Confidence: confidence,
			Language:   e.getLanguage(opts),
			Duration:   duration,
			Metadata:   opts.Metadata,
		}

		logger.Debug("OCR recognition completed",
			zap.Int("text_length", len(result.text)),
			zap.Float64("confidence", confidence),
			zap.Duration("duration", duration),
		)

		return ocrResult, nil
	}
}

// Close 关闭引擎
func (e *TesseractEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清理池中的客户端
	// sync.Pool 会自动垃圾回收，无需手动清理
	logger.Info("Tesseract engine closed")
	return nil
}

// GetSupportedLanguages 获取支持的语言
func (e *TesseractEngine) GetSupportedLanguages() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.supportedLanguages
}

// configureClient 配置客户端
func (e *TesseractEngine) configureClient(client *gosseract.Client, opts RecognizeOptions) error {
	// 设置语言
	lang := e.getLanguage(opts)
	if err := client.SetLanguage(lang); err != nil {
		return ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set language")
	}

	// 设置页面分割模式
	psm := e.config.PageSegMode
	if opts.PageSegMode != nil {
		psm = *opts.PageSegMode
	}
	if psm > 0 {
		if err := client.SetPageSegMode(gosseract.PageSegMode(psm)); err != nil {
			return ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set page seg mode")
		}
	}

	// 设置白名单
	if e.config.Whitelist != "" {
		client.SetWhitelist(e.config.Whitelist)
	}

	return nil
}

// getLanguage 获取要使用的语言
func (e *TesseractEngine) getLanguage(opts RecognizeOptions) string {
	if opts.Language != "" {
		return opts.Language
	}
	return e.config.Language
}

// getConfidence 获取置信度
func (e *TesseractEngine) getConfidence(client *gosseract.Client) float64 {
	// 尝试获取置信度，如果失败则返回 0
	conf, err := client.MeanConfidence()
	if err != nil {
		logger.Warn("Failed to get confidence", zap.Error(err))
		return 0
	}
	return float64(conf)
}

// RecognizeWithDetails 识别图像并返回详细信息(包含边界框)
func (e *TesseractEngine) RecognizeWithDetails(ctx context.Context, imageData []byte, opts RecognizeOptions) (*DetailedResult, error) {
	startTime := time.Now()

	client := e.clientPool.Get().(*gosseract.Client)
	defer e.clientPool.Put(client)

	// 应用配置
	if err := e.configureClient(client, opts); err != nil {
		return nil, err
	}

	// 设置图像数据
	if err := client.SetImageFromBytes(imageData); err != nil {
		return nil, ocrErrors.Wrap(err, ocrErrors.ErrOCREngineFailed, "failed to set image data")
	}

	// 执行 OCR
	resultChan := make(chan struct {
		boxes gosseract.BoundingBoxes
		err   error
	}, 1)

	go func() {
		boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
		resultChan <- struct {
			boxes gosseract.BoundingBoxes
			err   error
		}{boxes, err}
	}()

	// 等待结果或超时
	select {
	case <-ctx.Done():
		return nil, ocrErrors.New(ocrErrors.ErrTimeout, "OCR operation timeout")
	case result := <-resultChan:
		if result.err != nil {
			return nil, ocrErrors.Wrap(result.err, ocrErrors.ErrOCREngineFailed, "OCR recognition failed")
		}

		// 获取全文本
		text, _ := client.Text()

		// 转换边界框
		boundingBoxes := make([]BoundingBox, 0, len(result.boxes))
		var totalConf float64
		for _, box := range result.boxes {
			boundingBoxes = append(boundingBoxes, BoundingBox{
				X:      box.Box.Min.X,
				Y:      box.Box.Min.Y,
				Width:  box.Box.Max.X - box.Box.Min.X,
				Height: box.Box.Max.Y - box.Box.Min.Y,
				Text:   box.Word,
				Conf:   float64(box.Confidence),
			})
			totalConf += float64(box.Confidence)
		}

		// 计算平均置信度
		avgConf := 0.0
		if len(boundingBoxes) > 0 {
			avgConf = totalConf / float64(len(boundingBoxes))
		}

		duration := time.Since(startTime)

		return &DetailedResult{
			Text:        text,
			Confidence:  avgConf,
			BoundingBox: boundingBoxes,
			Duration:    duration,
		}, nil
	}
}

// ValidateLanguage 验证语言是否支持
func (e *TesseractEngine) ValidateLanguage(lang string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, supported := range e.supportedLanguages {
		if supported == lang {
			return nil
		}
	}

	return fmt.Errorf("unsupported language: %s", lang)
}