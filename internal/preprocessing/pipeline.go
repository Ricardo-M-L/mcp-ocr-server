package preprocessing

import (
	"context"
	"fmt"

	"gocv.io/x/gocv"
	"github.com/ricardo/mcp-ocr-server/internal/config"
)

// Processor 预处理器接口
type Processor interface {
	Process(ctx context.Context, input gocv.Mat) (gocv.Mat, error)
	Name() string
}

// Pipeline 智能预处理管道
type Pipeline struct {
	analyzer   *ImageAnalyzer
	config     config.PreprocessingConfig
	processors map[string]Processor
}

// NewPipeline 创建新的预处理管道
func NewPipeline(cfg config.PreprocessingConfig) *Pipeline {
	pipeline := &Pipeline{
		analyzer:   NewImageAnalyzer(),
		config:     cfg,
		processors: make(map[string]Processor),
	}

	// 注册所有处理器
	pipeline.processors["grayscale"] = NewGrayscaleProcessor()
	pipeline.processors["binarization"] = NewBinarizationProcessor(MethodOtsu, 0)
	pipeline.processors["denoise"] = NewDenoiseProcessor(DenoiseMedian, 5)
	pipeline.processors["deskew"] = NewDeskewProcessor(cfg.QualityThresholds.SkewAngleThreshold)

	return pipeline
}

// Process 执行智能预处理
func (p *Pipeline) Process(ctx context.Context, img gocv.Mat) (gocv.Mat, []string, *AnalysisResult, error) {
	if img.Empty() {
		return gocv.Mat{}, nil, nil, fmt.Errorf("input image is empty")
	}

	// 1. 分析图像质量
	analysis := p.analyzer.Analyze(img)

	// 2. 根据分析结果决定处理步骤
	steps := p.determineSteps(analysis)

	// 3. 执行处理步骤
	result := img.Clone()
	appliedSteps := []string{}

	for _, step := range steps {
		processor, ok := p.processors[step]
		if !ok {
			continue
		}

		// 执行处理
		processed, err := processor.Process(ctx, result)
		if err != nil {
			result.Close()
			return gocv.Mat{}, appliedSteps, analysis, fmt.Errorf("failed to process %s: %w", step, err)
		}

		// 关闭旧的结果
		if !result.Empty() {
			result.Close()
		}

		result = processed
		appliedSteps = append(appliedSteps, step)
	}

	return result, appliedSteps, analysis, nil
}

// determineSteps 根据图像分析结果决定处理步骤
func (p *Pipeline) determineSteps(analysis *AnalysisResult) []string {
	steps := []string{}

	// 总是先灰度化
	steps = append(steps, "grayscale")

	// 根据质量指标决定其他步骤
	thresholds := p.config.QualityThresholds

	// 如果噪声过高，进行去噪
	if analysis.NoiseLevel > thresholds.HighNoise {
		steps = append(steps, "denoise")
	}

	// 二值化（提高 OCR 准确度）
	steps = append(steps, "binarization")

	// 如果倾斜明显，进行校正
	if analysis.SkewAngle > thresholds.SkewAngleThreshold {
		steps = append(steps, "deskew")
	}

	return steps
}