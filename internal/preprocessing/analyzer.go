package preprocessing

import (
	"fmt"
	"math"

	"gocv.io/x/gocv"
)

// ImageQuality 图像质量评估结果
type ImageQuality struct {
	Sharpness  float64 // 清晰度 (越高越清晰)
	Contrast   float64 // 对比度 (0-255)
	Brightness float64 // 亮度 (0-255)
	NeedsPreprocessing bool // 是否需要预处理
	SuggestedPipeline []string // 建议的预处理步骤
}

// QualityAnalyzer 图像质量分析器
type QualityAnalyzer struct {
	sharpnessThreshold  float64
	contrastThreshold   float64
	brightnessMinThreshold float64
	brightnessMaxThreshold float64
}

// NewQualityAnalyzer 创建质量分析器
func NewQualityAnalyzer(sharpnessThreshold, contrastThreshold, brightnessMin float64) *QualityAnalyzer {
	return &QualityAnalyzer{
		sharpnessThreshold:     sharpnessThreshold,
		contrastThreshold:      contrastThreshold,
		brightnessMinThreshold: brightnessMin,
		brightnessMaxThreshold: 200.0, // 默认最大亮度阈值
	}
}

// Analyze 分析图像质量
func (a *QualityAnalyzer) Analyze(img gocv.Mat) (*ImageQuality, error) {
	if img.Empty() {
		return nil, fmt.Errorf("empty image")
	}

	quality := &ImageQuality{
		SuggestedPipeline: make([]string, 0),
	}

	// 转换为灰度图进行分析
	gray := gocv.NewMat()
	defer gray.Close()

	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		img.CopyTo(&gray)
	}

	// 1. 评估清晰度 (使用 Laplacian 方差)
	quality.Sharpness = a.calculateSharpness(gray)

	// 2. 评估对比度
	quality.Contrast = a.calculateContrast(gray)

	// 3. 评估亮度
	quality.Brightness = a.calculateBrightness(gray)

	// 4. 确定是否需要预处理
	quality.NeedsPreprocessing = a.determinePreprocessingNeed(quality)

	// 5. 生成建议的预处理管道
	quality.SuggestedPipeline = a.generatePipeline(quality)

	return quality, nil
}

// calculateSharpness 计算清晰度 (Laplacian 方差)
func (a *QualityAnalyzer) calculateSharpness(gray gocv.Mat) float64 {
	laplacian := gocv.NewMat()
	defer laplacian.Close()

	// 计算 Laplacian
	gocv.Laplacian(gray, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)

	// 计算方差
	mean := gocv.NewMat()
	stdDev := gocv.NewMat()
	defer mean.Close()
	defer stdDev.Close()

	gocv.MeanStdDev(laplacian, &mean, &stdDev)

	// 方差 = 标准差的平方
	variance := math.Pow(stdDev.GetDoubleAt(0, 0), 2)

	return variance
}

// calculateContrast 计算对比度 (标准差)
func (a *QualityAnalyzer) calculateContrast(gray gocv.Mat) float64 {
	mean := gocv.NewMat()
	stdDev := gocv.NewMat()
	defer mean.Close()
	defer stdDev.Close()

	gocv.MeanStdDev(gray, &mean, &stdDev)

	return stdDev.GetDoubleAt(0, 0)
}

// calculateBrightness 计算平均亮度
func (a *QualityAnalyzer) calculateBrightness(gray gocv.Mat) float64 {
	mean := gocv.Mean(gray)
	return mean.Val1
}

// determinePreprocessingNeed 判断是否需要预处理
func (a *QualityAnalyzer) determinePreprocessingNeed(quality *ImageQuality) bool {
	// 清晰度低
	if quality.Sharpness < a.sharpnessThreshold {
		return true
	}

	// 对比度低
	if quality.Contrast < a.contrastThreshold {
		return true
	}

	// 亮度不合适
	if quality.Brightness < a.brightnessMinThreshold || quality.Brightness > a.brightnessMaxThreshold {
		return true
	}

	return false
}

// generatePipeline 生成预处理管道
func (a *QualityAnalyzer) generatePipeline(quality *ImageQuality) []string {
	pipeline := make([]string, 0)

	// 始终转换为灰度
	pipeline = append(pipeline, "grayscale")

	// 亮度调整
	if quality.Brightness < a.brightnessMinThreshold {
		pipeline = append(pipeline, "brighten")
	} else if quality.Brightness > a.brightnessMaxThreshold {
		pipeline = append(pipeline, "darken")
	}

	// 对比度增强
	if quality.Contrast < a.contrastThreshold {
		pipeline = append(pipeline, "contrast_enhance")
	}

	// 降噪 (清晰度低时)
	if quality.Sharpness < a.sharpnessThreshold {
		pipeline = append(pipeline, "denoise")
	}

	// 二值化
	pipeline = append(pipeline, "binarization")

	// 倾斜校正
	pipeline = append(pipeline, "deskew")

	return pipeline
}

// CalculateSkewAngle 计算图像倾斜角度
func CalculateSkewAngle(img gocv.Mat) float64 {
	// 边缘检测
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(img, &edges, 50, 150)

	// 霍夫变换检测直线
	lines := gocv.NewMat()
	defer lines.Close()
	gocv.HoughLinesP(edges, &lines, 1, math.Pi/180, 100)

	if lines.Empty() || lines.Rows() < 10 {
		return 0.0
	}

	// 计算所有直线的角度
	angles := make([]float64, 0)
	for i := 0; i < lines.Rows(); i++ {
		x1 := float64(lines.GetIntAt(i, 0))
		y1 := float64(lines.GetIntAt(i, 1))
		x2 := float64(lines.GetIntAt(i, 2))
		y2 := float64(lines.GetIntAt(i, 3))

		angle := math.Atan2(y2-y1, x2-x1) * 180.0 / math.Pi

		// 只考虑接近水平的线
		if math.Abs(angle) < 45 {
			angles = append(angles, angle)
		}
	}

	if len(angles) == 0 {
		return 0.0
	}

	// 计算中位数角度
	return calculateMedian(angles)
}

// calculateMedian 计算中位数
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// 简单排序
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// 冒泡排序 (对于小数据集足够了)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2.0
	}
	return sorted[mid]
}