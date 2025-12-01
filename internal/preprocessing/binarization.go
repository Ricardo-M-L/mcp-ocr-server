package preprocessing

import (
	"context"

	"gocv.io/x/gocv"
)

// BinarizationMethod 二值化方法
type BinarizationMethod string

const (
	MethodOtsu      BinarizationMethod = "otsu"
	MethodAdaptive  BinarizationMethod = "adaptive"
	MethodThreshold BinarizationMethod = "threshold"
)

// BinarizationProcessor 二值化处理器
type BinarizationProcessor struct {
	method    BinarizationMethod
	threshold float64
}

// NewBinarizationProcessor 创建二值化处理器
func NewBinarizationProcessor(method BinarizationMethod, threshold float64) *BinarizationProcessor {
	if threshold == 0 {
		threshold = 127
	}
	return &BinarizationProcessor{
		method:    method,
		threshold: threshold,
	}
}

// Process 执行二值化处理
func (p *BinarizationProcessor) Process(ctx context.Context, input gocv.Mat) (gocv.Mat, error) {
	output := gocv.NewMat()

	switch p.method {
	case MethodOtsu:
		gocv.Threshold(input, &output, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	case MethodAdaptive:
		gocv.AdaptiveThreshold(input, &output, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 11, 2)
	default:
		gocv.Threshold(input, &output, p.threshold, 255, gocv.ThresholdBinary)
	}

	return output, nil
}

// Name 返回处理器名称
func (p *BinarizationProcessor) Name() string {
	return "binarization"
}