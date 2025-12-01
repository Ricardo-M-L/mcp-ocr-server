package errors

import "fmt"

// ErrorCode 定义错误代码类型
type ErrorCode string

const (
	ErrInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrFileNotFound        ErrorCode = "FILE_NOT_FOUND"
	ErrUnsupportedFormat   ErrorCode = "UNSUPPORTED_FORMAT"
	ErrImageTooLarge       ErrorCode = "IMAGE_TOO_LARGE"
	ErrPreprocessingFailed ErrorCode = "PREPROCESSING_FAILED"
	ErrOCREngineFailed     ErrorCode = "OCR_ENGINE_FAILED"
	ErrTimeout             ErrorCode = "TIMEOUT"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
)

// OCRError 自定义 OCR 错误类型
type OCRError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Err        error                  `json:"-"`
}

// Error 实现 error 接口
func (e *OCRError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现错误链
func (e *OCRError) Unwrap() error {
	return e.Err
}

// New 创建新的 OCR 错误
func New(code ErrorCode, message string) *OCRError {
	return &OCRError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *OCRError {
	return &OCRError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// WithDetails 添加详细信息
func (e *OCRError) WithDetails(key string, value interface{}) *OCRError {
	e.Details[key] = value
	return e
}