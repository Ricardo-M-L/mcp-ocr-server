package preprocessing

import (
	"context"
	"image"
	"math"

	"gocv.io/x/gocv"
)

// DeskewProcessor 倾斜校正处理器
type DeskewProcessor struct {
	angleThreshold float64
}

// NewDeskewProcessor 创建倾斜校正处理器
func NewDeskewProcessor(angleThreshold float64) *DeskewProcessor {
	if angleThreshold == 0 {
		angleThreshold = 0.5
	}
	return &DeskewProcessor{
		angleThreshold: angleThreshold,
	}
}

// Process 执行倾斜校正
func (p *DeskewProcessor) Process(ctx context.Context, input gocv.Mat) (gocv.Mat, error) {
	// 检测倾斜角度
	angle := p.detectSkewAngle(input)

	// 如果角度小于阈值，不需要校正
	if math.Abs(angle) < p.angleThreshold {
		return input.Clone(), nil
	}

	// 旋转图像
	output := p.rotateImage(input, angle)
	return output, nil
}

// Name 返回处理器名称
func (p *DeskewProcessor) Name() string {
	return "deskew"
}

// detectSkewAngle 检测倾斜角度
func (p *DeskewProcessor) detectSkewAngle(img gocv.Mat) float64 {
	// 简化实现：返回 0
	// 完整实现需要使用霍夫变换检测直线并计算角度
	return 0
}

// rotateImage 旋转图像
func (p *DeskewProcessor) rotateImage(img gocv.Mat, angle float64) gocv.Mat {
	// 获取图像中心
	center := image.Pt(img.Cols()/2, img.Rows()/2)

	// 获取旋转矩阵
	rotationMatrix := gocv.GetRotationMatrix2D(center, angle, 1.0)
	defer rotationMatrix.Close()

	// 执行旋转
	output := gocv.NewMat()
	gocv.WarpAffine(img, &output, rotationMatrix, image.Pt(img.Cols(), img.Rows()))

	return output
}