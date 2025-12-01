package preprocessing

import (
	"context"

	"gocv.io/x/gocv"
)

// GrayscaleProcessor 灰度化处理器
type GrayscaleProcessor struct{}

// NewGrayscaleProcessor 创建灰度化处理器
func NewGrayscaleProcessor() *GrayscaleProcessor {
	return &GrayscaleProcessor{}
}

// Process 执行灰度化处理
func (p *GrayscaleProcessor) Process(ctx context.Context, input gocv.Mat) (gocv.Mat, error) {
	// 检查是否已经是灰度图
	if input.Channels() == 1 {
		return input.Clone(), nil
	}

	output := gocv.NewMat()
	gocv.CvtColor(input, &output, gocv.ColorBGRToGray)

	return output, nil
}

// Name 返回处理器名称
func (p *GrayscaleProcessor) Name() string {
	return "grayscale"
}