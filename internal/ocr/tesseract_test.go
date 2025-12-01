package ocr

import (
	"context"
	"testing"
	"time"
)

func TestTesseractEngine_Init(t *testing.T) {
	engine := NewTesseractEngine()

	config := EngineConfig{
		Language:    "eng",
		PageSegMode: 3,
		Timeout:     time.Second * 30,
	}

	err := engine.Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}

	defer engine.Close()
}

func TestTesseractEngine_GetSupportedLanguages(t *testing.T) {
	engine := NewTesseractEngine()

	config := EngineConfig{
		Language: "eng",
	}

	err := engine.Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	languages := engine.GetSupportedLanguages()
	if len(languages) == 0 {
		t.Error("Expected supported languages, got empty list")
	}

	t.Logf("Supported languages: %v", languages)
}

func TestTesseractEngine_RecognizeText(t *testing.T) {
	// 这个测试需要实际的图像数据
	// 在实际测试中，你需要准备测试图像
	t.Skip("Skipping integration test - requires test image")

	engine := NewTesseractEngine()

	config := EngineConfig{
		Language:    "eng",
		PageSegMode: 3,
		Timeout:     time.Second * 30,
	}

	err := engine.Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	// 模拟图像数据
	// imageData := []byte{...}

	ctx := context.Background()
	opts := RecognizeOptions{
		Language:   "eng",
		Preprocess: false,
	}

	// result, err := engine.RecognizeText(ctx, imageData, opts)
	// if err != nil {
	// 	t.Fatalf("Failed to recognize text: %v", err)
	// }

	// t.Logf("Recognized text: %s", result.Text)
	// t.Logf("Confidence: %.2f", result.Confidence)
}