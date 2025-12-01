package preprocessing

import (
	"context"
	"image"

	"gocv.io/x/gocv"
)

// DenoiseMethod 去噪方法
type DenoiseMethod string

const (
	DenoiseMedian     DenoiseMethod = "median"
	DenoiseBilateral  DenoiseMethod = "bilateral"
	DenoiseFastNl     DenoiseMethod = "fastNlMeans"
)

// DenoiseProcessor 去噪处理器
type DenoiseProcessor struct {
	method     DenoiseMethod
	kernelSize int
}

// NewDenoiseProcessor 创建去噪处理器
func NewDenoiseProcessor(method DenoiseMethod, kernelSize int) *DenoiseProcessor {
	if kernelSize == 0 {
		kernelSize = 5
	}
	// 确保 kernel size 是奇数
	if kernelSize%2 == 0 {
		kernelSize++
	}
	return &DenoiseProcessor{
		method:     method,
		kernelSize: kernelSize,
	}
}

// Process 执行去噪处理
func (p *DenoiseProcessor) Process(ctx context.Context, input gocv.Mat) (gocv.Mat, error) {
	output := gocv.NewMat()

	switch p.method {
	case DenoiseMedian:
		gocv.MedianBlur(input, &output, p.kernelSize)
	case DenoiseBilateral:
		gocv.BilateralFilter(input, &output, p.kernelSize, 75, 75)
	case DenoiseFastNl:
		if input.Channels() == 1 {
			gocv.FastNlMeansDenoising(input, &output)
		} else {
			gocv.FastNlMeansDenoisingColored(input, &output)
		}
	default:
		gocv.MedianBlur(input, &output, p.kernelSize)
	}

	return output, nil
}

// Name 返回处理器名称
func (p *DenoiseProcessor) Name() string {
	return "denoise"
}