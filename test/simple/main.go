package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ricardo/mcp-ocr-server/internal/ocr"
)

// 这是一个简单的测试程序，用于验证 OCR 引擎是否正常工作
// 使用方法: go run test/simple/main.go <image_path>

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test/simple/main.go <image_path>")
		os.Exit(1)
	}

	imagePath := os.Args[1]

	fmt.Printf("Testing OCR with image: %s\n", imagePath)

	// 读取图像
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Failed to read image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Image size: %d bytes\n", len(imageData))

	// 创建 OCR 引擎
	engine := ocr.NewTesseractEngine()

	config := ocr.EngineConfig{
		Language:    "eng",
		PageSegMode: 3,
		Timeout:     time.Second * 30,
	}

	fmt.Println("Initializing OCR engine...")
	if err := engine.Init(config); err != nil {
		fmt.Printf("Failed to initialize engine: %v\n", err)
		os.Exit(1)
	}
	defer engine.Close()

	// 执行 OCR
	fmt.Println("Recognizing text...")
	ctx := context.Background()
	opts := ocr.RecognizeOptions{
		Language:   "eng",
		Preprocess: false,
	}

	result, err := engine.RecognizeText(ctx, imageData, opts)
	if err != nil {
		fmt.Printf("OCR failed: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	fmt.Println("\n=== OCR Result ===")
	fmt.Printf("Text:\n%s\n", result.Text)
	fmt.Printf("\nConfidence: %.2f%%\n", result.Confidence)
	fmt.Printf("Language: %s\n", result.Language)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Println("==================")
}