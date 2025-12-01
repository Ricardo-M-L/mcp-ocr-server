.PHONY: help deps build run test clean install docker-build docker-run

# 变量定义
APP_NAME := mcp-ocr-server
VERSION := 1.0.0
BUILD_DIR := bin
CMD_DIR := cmd/server
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.appVersion=$(VERSION)"

# 默认目标
help:
	@echo "MCP OCR Server - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make deps          - Install Go dependencies"
	@echo "  make build         - Build the server binary"
	@echo "  make run           - Run the server"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install system dependencies (Tesseract, OpenCV)"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run in Docker container"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo ""

# 安装 Go 依赖
deps:
	@echo "Installing Go dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✓ Dependencies installed"

# 编译
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "✓ Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run: build
	@echo "Starting $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME) -config configs/config.yaml

# 测试
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

# 测试覆盖率
coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "✓ Clean complete"

# 安装系统依赖
install:
	@echo "Installing system dependencies..."
	./scripts/install-deps.sh

# 代码格式化
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "✓ Format complete"

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		exit 1; \
	fi

# 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "✓ Docker image built: $(APP_NAME):$(VERSION)"

# 运行 Docker 容器
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it \
		-v $(PWD)/configs:/app/configs \
		-v $(PWD)/test:/app/test \
		$(APP_NAME):latest

# 开发模式运行 (带热重载)
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with:"; \
		echo "  go install github.com/cosmtrek/air@latest"; \
		$(MAKE) run; \
	fi

# 生成模拟数据
mock:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/ocr/engine.go -destination=internal/ocr/mock_engine.go -package=ocr; \
		echo "✓ Mocks generated"; \
	else \
		echo "mockgen not installed. Install with:"; \
		echo "  go install github.com/golang/mock/mockgen@latest"; \
	fi

# 验证依赖
verify-deps:
	@echo "Verifying dependencies..."
	@command -v tesseract >/dev/null 2>&1 || { echo "✗ Tesseract not found"; exit 1; }
	@tesseract --version
	@echo "✓ Dependencies verified"

# 完整构建 (包括依赖检查和测试)
all: verify-deps deps test build
	@echo "✓ All tasks complete"