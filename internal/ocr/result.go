package ocr

import "time"

// Result OCR 识别结果
type Result struct {
	Text       string       `json:"text"`
	Confidence float64      `json:"confidence"`
	Words      []WordDetail `json:"words"`
	Language   string       `json:"language"`
	Duration   time.Duration `json:"duration"`
}

// WordDetail 单词详细信息
type WordDetail struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	BBox       BBox    `json:"bbox"`
}

// BBox 边界框
type BBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Options OCR 选项
type Options struct {
	Language          string
	PSM               int
	ConfidenceThreshold float64
}