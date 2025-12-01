package preprocessing

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
	ocrErrors "github.com/ricardo/mcp-ocr-server/pkg/errors"
	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// Config 预处理配置
type Config struct {
	Enabled           bool
	AutoMode          bool
	Grayscale         bool
	Denoise           bool
	DenoiseStrength   int
	Binarization      bool
	BinarizationMode  string // "otsu" or "adaptive"
	AdaptiveBlockSize int
	AdaptiveC         float64
	DeskewCorrection  bool
	DeskewAngleLimit  float64
	Resize            bool
	ResizeWidth       int
	ResizeHeight      int
	QualityThresholds struct {
		Sharpness  float64
		Contrast   float64
		Brightness float64
	}
}

// Preprocessor 图像预处理器
type Preprocessor struct {
	config  Config
	analyzer *QualityAnalyzer
}

// NewPreprocessor 创建预处理器
func NewPreprocessor(config Config) *Preprocessor {
	analyzer := NewQualityAnalyzer(
		config.QualityThresholds.Sharpness,
		config.QualityThresholds.Contrast,
		config.QualityThresholds.Brightness,
	)

	return &Preprocessor{
		config:   config,
		analyzer: analyzer,
	}
}

// Process 处理图像
func (p *Preprocessor) Process(imageData []byte) ([]byte, error) {
	if !p.config.Enabled {
		return imageData, nil
	}

	// 解码图像
	img, err := gocv.IMDecode(imageData, gocv.IMReadColor)
	if err != nil {
		return nil, ocrErrors.Wrap(err, ocrErrors.ErrPreprocessingFailed, "failed to decode image")
	}
	defer img.Close()

	if img.Empty() {
		return nil, ocrErrors.New(ocrErrors.ErrPreprocessingFailed, "decoded image is empty")
	}

	logger.Debug("Image loaded",
		zap.Int("width", img.Cols()),
		zap.Int("height", img.Rows()),
		zap.Int("channels", img.Channels()),
	)

	// 自动模式：分析图像质量
	var pipeline []string
	if p.config.AutoMode {
		quality, err := p.analyzer.Analyze(img)
		if err != nil {
			logger.Warn("Failed to analyze image quality", zap.Error(err))
			pipeline = p.getDefaultPipeline()
		} else {
			logger.Info("Image quality analysis",
				zap.Float64("sharpness", quality.Sharpness),
				zap.Float64("contrast", quality.Contrast),
				zap.Float64("brightness", quality.Brightness),
				zap.Bool("needs_preprocessing", quality.NeedsPreprocessing),
			)
			pipeline = quality.SuggestedPipeline
		}
	} else {
		pipeline = p.getDefaultPipeline()
	}

	logger.Info("Preprocessing pipeline", zap.Strings("steps", pipeline))

	// 执行预处理管道
	processed := img.Clone()
	defer processed.Close()

	for _, step := range pipeline {
		var err error
		processed, err = p.applyStep(processed, step)
		if err != nil {
			return nil, ocrErrors.Wrap(err, ocrErrors.ErrPreprocessingFailed, fmt.Sprintf("preprocessing step '%s' failed", step))
		}
	}

	// 编码为 PNG
	buf, err := gocv.IMEncode(".png", processed)
	if err != nil {
		return nil, ocrErrors.Wrap(err, ocrErrors.ErrPreprocessingFailed, "failed to encode processed image")
	}

	result := buf.GetBytes()
	buf.Close()

	logger.Debug("Image preprocessing completed", zap.Int("output_size", len(result)))

	return result, nil
}

// applyStep 应用单个预处理步骤
func (p *Preprocessor) applyStep(img gocv.Mat, step string) (gocv.Mat, error) {
	result := gocv.NewMat()

	switch step {
	case "grayscale":
		if img.Channels() > 1 {
			gocv.CvtColor(img, &result, gocv.ColorBGRToGray)
		} else {
			img.CopyTo(&result)
		}

	case "denoise":
		if p.config.Denoise {
			// 使用 fastNlMeansDenoising
			if img.Channels() == 1 {
				gocv.FastNlMeansDenoising(img, &result)
			} else {
				gocv.FastNlMeansDenoisingColored(img, &result)
			}
		} else {
			img.CopyTo(&result)
		}

	case "binarization":
		if p.config.Binarization {
			result = p.applyBinarization(img)
		} else {
			img.CopyTo(&result)
		}

	case "deskew":
		if p.config.DeskewCorrection {
			result = p.applyDeskew(img)
		} else {
			img.CopyTo(&result)
		}

	case "contrast_enhance":
		result = p.enhanceContrast(img)

	case "brighten":
		result = p.adjustBrightness(img, 30)

	case "darken":
		result = p.adjustBrightness(img, -30)

	case "resize":
		if p.config.Resize {
			result = p.applyResize(img)
		} else {
			img.CopyTo(&result)
		}

	default:
		img.CopyTo(&result)
	}

	// 释放原图像
	if !img.Empty() {
		img.Close()
	}

	return result, nil
}

// applyBinarization 应用二值化
func (p *Preprocessor) applyBinarization(img gocv.Mat) gocv.Mat {
	result := gocv.NewMat()

	// 确保是灰度图
	gray := gocv.NewMat()
	defer gray.Close()

	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		img.CopyTo(&gray)
	}

	switch p.config.BinarizationMode {
	case "otsu":
		gocv.Threshold(gray, &result, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	case "adaptive":
		gocv.AdaptiveThreshold(
			gray,
			&result,
			255,
			gocv.AdaptiveThresholdMean,
			gocv.ThresholdBinary,
			p.config.AdaptiveBlockSize,
			p.config.AdaptiveC,
		)
	default:
		gocv.Threshold(gray, &result, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	}

	return result
}

// applyDeskew 应用倾斜校正
func (p *Preprocessor) applyDeskew(img gocv.Mat) gocv.Mat {
	// 确保是灰度图
	gray := gocv.NewMat()
	defer gray.Close()

	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		img.CopyTo(&gray)
	}

	// 计算倾斜角度
	angle := CalculateSkewAngle(gray)

	logger.Debug("Detected skew angle", zap.Float64("angle", angle))

	// 如果角度在限制范围内，进行校正
	if math.Abs(angle) > 0.5 && math.Abs(angle) < p.config.DeskewAngleLimit {
		return p.rotateImage(img, angle)
	}

	result := gocv.NewMat()
	img.CopyTo(&result)
	return result
}

// rotateImage 旋转图像
func (p *Preprocessor) rotateImage(img gocv.Mat, angle float64) gocv.Mat {
	center := image.Point{X: img.Cols() / 2, Y: img.Rows() / 2}
	rotMatrix := gocv.GetRotationMatrix2D(center, angle, 1.0)
	defer rotMatrix.Close()

	result := gocv.NewMat()
	gocv.WarpAffine(img, &result, rotMatrix, image.Point{X: img.Cols(), Y: img.Rows()})

	return result
}

// enhanceContrast 增强对比度
func (p *Preprocessor) enhanceContrast(img gocv.Mat) gocv.Mat {
	result := gocv.NewMat()

	// 使用 CLAHE (对比度受限自适应直方图均衡化)
	clahe := gocv.NewCLAHE()
	clahe.SetClipLimit(2.0)
	clahe.SetTilesGridSize(image.Point{X: 8, Y: 8})
	defer clahe.Close()

	if img.Channels() == 1 {
		clahe.Apply(img, &result)
	} else {
		// 转换到 LAB 色彩空间
		lab := gocv.NewMat()
		defer lab.Close()
		gocv.CvtColor(img, &lab, gocv.ColorBGRToLab)

		// 分离通道
		channels := gocv.Split(lab)
		defer func() {
			for _, ch := range channels {
				ch.Close()
			}
		}()

		// 只对 L 通道应用 CLAHE
		lEnhanced := gocv.NewMat()
		defer lEnhanced.Close()
		clahe.Apply(channels[0], &lEnhanced)

		// 合并通道
		lEnhanced.CopyTo(&channels[0])
		gocv.Merge(channels, &lab)

		// 转换回 BGR
		gocv.CvtColor(lab, &result, gocv.ColorLabToBGR)
	}

	return result
}

// adjustBrightness 调整亮度
func (p *Preprocessor) adjustBrightness(img gocv.Mat, delta int) gocv.Mat {
	result := gocv.NewMat()
	img.ConvertTo(&result, -1, 1.0, float64(delta))
	return result
}

// applyResize 调整图像大小
func (p *Preprocessor) applyResize(img gocv.Mat) gocv.Mat {
	result := gocv.NewMat()

	width := p.config.ResizeWidth
	height := p.config.ResizeHeight

	// 如果只指定了一个维度，保持宽高比
	if width > 0 && height == 0 {
		ratio := float64(width) / float64(img.Cols())
		height = int(float64(img.Rows()) * ratio)
	} else if height > 0 && width == 0 {
		ratio := float64(height) / float64(img.Rows())
		width = int(float64(img.Cols()) * ratio)
	}

	if width > 0 && height > 0 {
		gocv.Resize(img, &result, image.Point{X: width, Y: height}, 0, 0, gocv.InterpolationLinear)
	} else {
		img.CopyTo(&result)
	}

	return result
}

// getDefaultPipeline 获取默认预处理管道
func (p *Preprocessor) getDefaultPipeline() []string {
	pipeline := make([]string, 0)

	if p.config.Grayscale {
		pipeline = append(pipeline, "grayscale")
	}

	if p.config.Denoise {
		pipeline = append(pipeline, "denoise")
	}

	if p.config.Binarization {
		pipeline = append(pipeline, "binarization")
	}

	if p.config.DeskewCorrection {
		pipeline = append(pipeline, "deskew")
	}

	if p.config.Resize {
		pipeline = append(pipeline, "resize")
	}

	return pipeline
}