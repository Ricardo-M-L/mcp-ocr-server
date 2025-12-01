package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ricardo/mcp-ocr-server/internal/cache"
	"github.com/ricardo/mcp-ocr-server/internal/config"
	"github.com/ricardo/mcp-ocr-server/internal/ocr"
	"github.com/ricardo/mcp-ocr-server/internal/pool"
	"github.com/ricardo/mcp-ocr-server/internal/preprocessing"
	ocrErrors "github.com/ricardo/mcp-ocr-server/pkg/errors"
	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// Handler OCR Tool Handler
type Handler struct {
	engine       ocr.Engine
	preprocessor *preprocessing.Preprocessor
	cache        *cache.Cache
	workerPool   *pool.WorkerPool
	config       *config.Config
}

// NewHandler 创建 Tool Handler
func NewHandler(cfg *config.Config) (*Handler, error) {
	// 创建 OCR 引擎
	engine := ocr.NewTesseractEngine()
	engineConfig := ocr.EngineConfig{
		Language:    cfg.OCR.Language,
		DataPath:    cfg.OCR.DataPath,
		PageSegMode: cfg.OCR.PageSegMode,
		EngineMode:  cfg.OCR.EngineMode,
		Whitelist:   cfg.OCR.Whitelist,
		Timeout:     time.Duration(cfg.OCR.Timeout) * time.Second,
	}

	if err := engine.Init(engineConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize OCR engine: %w", err)
	}

	// 创建预处理器
	preprocessorConfig := preprocessing.Config{
		Enabled:           cfg.Preprocessing.Enabled,
		AutoMode:          cfg.Preprocessing.AutoMode,
		Grayscale:         cfg.Preprocessing.Grayscale,
		Denoise:           cfg.Preprocessing.Denoise,
		DenoiseStrength:   cfg.Preprocessing.DenoiseStrength,
		Binarization:      cfg.Preprocessing.Binarization,
		BinarizationMode:  cfg.Preprocessing.BinarizationMode,
		AdaptiveBlockSize: cfg.Preprocessing.AdaptiveBlockSize,
		AdaptiveC:         cfg.Preprocessing.AdaptiveC,
		DeskewCorrection:  cfg.Preprocessing.DeskewCorrection,
		DeskewAngleLimit:  cfg.Preprocessing.DeskewAngleLimit,
		Resize:            cfg.Preprocessing.Resize,
		ResizeWidth:       cfg.Preprocessing.ResizeWidth,
		ResizeHeight:      cfg.Preprocessing.ResizeHeight,
	}
	preprocessorConfig.QualityThresholds.Sharpness = cfg.Preprocessing.QualityThresholds.Sharpness
	preprocessorConfig.QualityThresholds.Contrast = cfg.Preprocessing.QualityThresholds.Contrast
	preprocessorConfig.QualityThresholds.Brightness = cfg.Preprocessing.QualityThresholds.Brightness

	preprocessor := preprocessing.NewPreprocessor(preprocessorConfig)

	// 创建缓存
	cacheTTL := time.Duration(cfg.Performance.CacheTTL) * time.Second
	resultCache := cache.NewCache(cfg.Performance.CacheSize, cacheTTL, cfg.Performance.CacheEnabled)

	// 创建 Worker Pool
	workerPool := pool.NewWorkerPool(cfg.Performance.WorkerPoolSize, cfg.Performance.QueueSize)
	if err := workerPool.Start(); err != nil {
		return nil, fmt.Errorf("failed to start worker pool: %w", err)
	}

	return &Handler{
		engine:       engine,
		preprocessor: preprocessor,
		cache:        resultCache,
		workerPool:   workerPool,
		config:       cfg,
	}, nil
}

// Handle 处理工具调用
func (h *Handler) Handle(ctx context.Context, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	logger.Info("Tool called", zap.String("tool", toolName))

	switch toolName {
	case "ocr_recognize_text":
		return h.handleRecognizeText(ctx, arguments)
	case "ocr_recognize_text_base64":
		return h.handleRecognizeTextBase64(ctx, arguments)
	case "ocr_batch_recognize":
		return h.handleBatchRecognize(ctx, arguments)
	case "ocr_get_supported_languages":
		return h.handleGetSupportedLanguages(ctx, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// handleRecognizeText 处理文本识别
func (h *Handler) handleRecognizeText(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// 解析参数
	imagePath, ok := args["image_path"].(string)
	if !ok {
		return nil, ocrErrors.New(ocrErrors.ErrInvalidInput, "image_path is required")
	}

	language := h.getStringArg(args, "language", h.config.OCR.Language)
	preprocess := h.getBoolArg(args, "preprocess", true)
	autoMode := h.getBoolArg(args, "auto_mode", true)

	// 读取图像文件
	imageData, err := h.readImageFile(imagePath)
	if err != nil {
		return h.errorResult(err), nil
	}

	// 执行 OCR
	result, err := h.recognizeImage(ctx, imageData, language, preprocess, autoMode)
	if err != nil {
		return h.errorResult(err), nil
	}

	return h.successResult(result), nil
}

// handleRecognizeTextBase64 处理 Base64 图像识别
func (h *Handler) handleRecognizeTextBase64(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// 解析参数
	imageBase64, ok := args["image_base64"].(string)
	if !ok {
		return nil, ocrErrors.New(ocrErrors.ErrInvalidInput, "image_base64 is required")
	}

	language := h.getStringArg(args, "language", h.config.OCR.Language)
	preprocess := h.getBoolArg(args, "preprocess", true)
	autoMode := h.getBoolArg(args, "auto_mode", true)

	// 解码 Base64
	imageData, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return h.errorResult(ocrErrors.Wrap(err, ocrErrors.ErrInvalidInput, "invalid base64 data")), nil
	}

	// 执行 OCR
	result, err := h.recognizeImage(ctx, imageData, language, preprocess, autoMode)
	if err != nil {
		return h.errorResult(err), nil
	}

	return h.successResult(result), nil
}

// handleBatchRecognize 处理批量识别
func (h *Handler) handleBatchRecognize(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// 解析参数
	imagePathsInterface, ok := args["image_paths"].([]interface{})
	if !ok {
		return nil, ocrErrors.New(ocrErrors.ErrInvalidInput, "image_paths is required")
	}

	imagePaths := make([]string, 0, len(imagePathsInterface))
	for _, p := range imagePathsInterface {
		if path, ok := p.(string); ok {
			imagePaths = append(imagePaths, path)
		}
	}

	if len(imagePaths) == 0 {
		return nil, ocrErrors.New(ocrErrors.ErrInvalidInput, "no valid image paths provided")
	}

	language := h.getStringArg(args, "language", h.config.OCR.Language)
	preprocess := h.getBoolArg(args, "preprocess", true)
	autoMode := h.getBoolArg(args, "auto_mode", true)

	// 并行处理
	results := make([]map[string]interface{}, len(imagePaths))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, path := range imagePaths {
		wg.Add(1)
		go func(index int, imagePath string) {
			defer wg.Done()

			imageData, err := h.readImageFile(imagePath)
			if err != nil {
				mu.Lock()
				results[index] = map[string]interface{}{
					"path":  imagePath,
					"error": err.Error(),
				}
				mu.Unlock()
				return
			}

			result, err := h.recognizeImage(ctx, imageData, language, preprocess, autoMode)
			if err != nil {
				mu.Lock()
				results[index] = map[string]interface{}{
					"path":  imagePath,
					"error": err.Error(),
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			resultMap := map[string]interface{}{
				"path":       imagePath,
				"text":       result.Text,
				"confidence": result.Confidence,
				"language":   result.Language,
				"duration":   result.Duration.Seconds(),
			}
			results[index] = resultMap
			mu.Unlock()
		}(i, path)
	}

	wg.Wait()

	return h.successResult(map[string]interface{}{
		"results": results,
		"count":   len(results),
	}), nil
}

// handleGetSupportedLanguages 获取支持的语言
func (h *Handler) handleGetSupportedLanguages(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	languages := h.engine.GetSupportedLanguages()

	return h.successResult(map[string]interface{}{
		"languages": languages,
	}), nil
}

// recognizeImage 识别图像
func (h *Handler) recognizeImage(ctx context.Context, imageData []byte, language string, preprocess, autoMode bool) (*ocr.RecognizeResult, error) {
	// 检查图像大小
	if int64(len(imageData)) > h.config.OCR.MaxImageSize {
		return nil, ocrErrors.New(ocrErrors.ErrImageTooLarge, fmt.Sprintf("image size exceeds limit: %d bytes", len(imageData)))
	}

	// 生成缓存键
	cacheKey := cache.GenerateKey(imageData, language, fmt.Sprintf("%t", preprocess))

	// 检查缓存
	if cached, found := h.cache.Get(cacheKey); found {
		if result, ok := cached.(*ocr.RecognizeResult); ok {
			logger.Info("OCR result from cache", zap.String("language", language))
			return result, nil
		}
	}

	// 预处理
	processedData := imageData
	if preprocess {
		var err error
		processedData, err = h.preprocessor.Process(imageData)
		if err != nil {
			logger.Warn("Preprocessing failed, using original image", zap.Error(err))
			processedData = imageData
		}
	}

	// 执行 OCR
	opts := ocr.RecognizeOptions{
		Language:   language,
		Preprocess: preprocess,
		Metadata: map[string]string{
			"auto_mode": fmt.Sprintf("%t", autoMode),
		},
	}

	result, err := h.engine.RecognizeText(ctx, processedData, opts)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	h.cache.Set(cacheKey, result)

	return result, nil
}

// readImageFile 读取图像文件
func (h *Handler) readImageFile(path string) ([]byte, error) {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查文件是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return nil, ocrErrors.New(ocrErrors.ErrFileNotFound, fmt.Sprintf("file not found: %s", path))
	}

	// 读取文件
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, ocrErrors.Wrap(err, ocrErrors.ErrInternalError, "failed to read image file")
	}

	return data, nil
}

// successResult 创建成功结果
func (h *Handler) successResult(data interface{}) *mcp.CallToolResult {
	jsonData, _ := json.Marshal(data)

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}
}

// errorResult 创建错误结果
func (h *Handler) errorResult(err error) *mcp.CallToolResult {
	errData := map[string]interface{}{
		"error": err.Error(),
	}

	if ocrErr, ok := err.(*ocrErrors.OCRError); ok {
		errData["code"] = string(ocrErr.Code)
		errData["details"] = ocrErr.Details
	}

	jsonData, _ := json.Marshal(errData)

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		IsError: true,
	}
}

// getStringArg 获取字符串参数
func (h *Handler) getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

// getBoolArg 获取布尔参数
func (h *Handler) getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}

// Close 关闭 Handler
func (h *Handler) Close() error {
	h.workerPool.Stop()
	h.cache.Clear()
	return h.engine.Close()
}