package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetToolSchemas 获取所有 MCP Tool Schema
func GetToolSchemas() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "ocr_recognize_text",
			Description: "Recognize text from an image using OCR with intelligent preprocessing",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"image_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the image file to process",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language for OCR recognition (eng, chi_sim, chi_tra, jpn, or combination like 'eng+chi_sim')",
						"default":     "eng",
					},
					"preprocess": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable image preprocessing for better OCR results",
						"default":     true,
					},
					"auto_mode": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable automatic quality analysis and adaptive preprocessing",
						"default":     true,
					},
				},
				Required: []string{"image_path"},
			},
		},
		{
			Name:        "ocr_recognize_text_base64",
			Description: "Recognize text from a base64-encoded image using OCR",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"image_base64": map[string]interface{}{
						"type":        "string",
						"description": "Base64-encoded image data",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language for OCR recognition",
						"default":     "eng",
					},
					"preprocess": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable image preprocessing",
						"default":     true,
					},
					"auto_mode": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable automatic quality analysis",
						"default":     true,
					},
				},
				Required: []string{"image_base64"},
			},
		},
		{
			Name:        "ocr_batch_recognize",
			Description: "Recognize text from multiple images in batch",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"image_paths": map[string]interface{}{
						"type":        "array",
						"description": "Array of image file paths to process",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language for OCR recognition",
						"default":     "eng",
					},
					"preprocess": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable image preprocessing",
						"default":     true,
					},
					"auto_mode": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable automatic quality analysis",
						"default":     true,
					},
				},
				Required: []string{"image_paths"},
			},
		},
		{
			Name:        "ocr_get_supported_languages",
			Description: "Get list of supported OCR languages",
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
	}
}