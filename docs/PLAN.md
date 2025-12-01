# MCP OCR Server 实施计划

## 项目概述

**项目名称**: MCP OCR Server
**项目路径**: `/Users/ricardo/Documents/公司学习文件/自己开发的mcp`
**开发语言**: Go 1.23+
**项目定位**: 生产级图片 OCR 文字提取 MCP 工具

## 核心技术栈

### 基础依赖
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` v1.1.0+
- **OCR 引擎 1**: `github.com/otiai10/gosseract/v2` (Tesseract CGo 绑定)
- **OCR 引擎 2**: Gogosseract (Tesseract WASM 版本)
- **图像处理**: `gocv.io/x/gocv` v0.38.0+ (OpenCV 4 Go 绑定)
- **日志**: `go.uber.org/zap` v1.27.0+
- **配置**: `gopkg.in/yaml.v3` v3.0.1+

### 语言支持
- 英文 (eng)
- 简体中文 (chi_sim)
- 繁体中文 (chi_tra)
- 日文 (jpn)

## 项目架构

### 目录结构

```
mcp-ocr-server/
├── cmd/
│   └── server/
│       └── main.go                 # 服务启动入口
├── internal/
│   ├── config/
│   │   ├── config.go              # 配置结构定义
│   │   └── loader.go              # 配置加载逻辑
│   ├── ocr/
│   │   ├── engine.go              # OCR 引擎接口
│   │   ├── tesseract_cgo.go       # Gosseract 实现
│   │   ├── tesseract_wasm.go      # Gogosseract 实现
│   │   ├── pool.go                # 引擎资源池
│   │   └── result.go              # OCR 结果结构
│   ├── preprocessing/
│   │   ├── pipeline.go            # 智能预处理管道
│   │   ├── analyzer.go            # 图像质量分析器
│   │   ├── grayscale.go           # 灰度化
│   │   ├── binarization.go        # 二值化
│   │   ├── denoise.go             # 去噪
│   │   └── deskew.go              # 倾斜校正
│   ├── tools/
│   │   ├── ocr_tool.go            # MCP OCR Tool
│   │   └── schemas.go             # Input/Output Schema
│   ├── server/
│   │   ├── server.go              # MCP Server 封装
│   │   └── handlers.go            # 工具处理器
│   ├── cache/
│   │   └── cache.go               # 结果缓存
│   └── pool/
│       └── worker_pool.go         # Worker Pool
├── pkg/
│   ├── errors/
│   │   └── errors.go              # 自定义错误
│   └── logger/
│       └── logger.go              # 日志封装
├── configs/
│   ├── config.yaml                # 默认配置
│   └── config.example.yaml        # 配置示例
├── scripts/
│   ├── install_deps.sh            # 依赖安装
│   └── build.sh                   # 构建脚本
├── test/
│   ├── testdata/                  # 测试图片
│   └── integration/               # 集成测试
├── docs/
│   ├── README.md                  # 项目文档
│   ├── ARCHITECTURE.md            # 架构设计
│   ├── API.md                     # API 文档
│   └── DEPLOYMENT.md              # 部署指南
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── .gitignore
```

### 核心模块设计

#### 1. OCR 引擎接口 (`internal/ocr/engine.go`)

```go
type OCREngine interface {
    // 初始化引擎
    Initialize(config Config) error

    // 从图像提取文字
    ExtractText(ctx context.Context, imagePath string, opts Options) (*Result, error)

    // 从字节数组提取文字
    ExtractTextFromBytes(ctx context.Context, imageData []byte, opts Options) (*Result, error)

    // 获取引擎类型
    Type() EngineType

    // 关闭引擎
    Close() error
}

type EngineType string

const (
    EngineTypeCGo  EngineType = "cgo"
    EngineTypeWASM EngineType = "wasm"
)
```

#### 2. 智能预处理管道 (`internal/preprocessing/pipeline.go`)

```go
// 预处理分析器
type ImageAnalyzer struct{}

func (a *ImageAnalyzer) Analyze(img gocv.Mat) *AnalysisResult {
    return &AnalysisResult{
        Brightness:  a.calculateBrightness(img),
        Contrast:    a.calculateContrast(img),
        Sharpness:   a.calculateSharpness(img),
        SkewAngle:   a.detectSkewAngle(img),
        NoiseLevel:  a.estimateNoiseLevel(img),
    }
}

// 智能预处理管道
type IntelligentPipeline struct {
    analyzer *ImageAnalyzer
}

func (p *IntelligentPipeline) Process(ctx context.Context, img gocv.Mat) (gocv.Mat, []string, error) {
    // 1. 分析图像质量
    analysis := p.analyzer.Analyze(img)

    // 2. 根据分析结果决定处理步骤
    steps := p.determineSteps(analysis)

    // 3. 执行处理步骤
    result, appliedSteps := p.applySteps(ctx, img, steps)

    return result, appliedSteps, nil
}
```

#### 3. 引擎资源池 (`internal/ocr/pool.go`)

```go
type EnginePool struct {
    cgoEngines  chan *TesseractCGoEngine
    wasmEngines chan *TesseractWASMEngine
    maxSize     int
    factory     EngineFactory
}

func (ep *EnginePool) Get(engineType EngineType) (OCREngine, error) {
    switch engineType {
    case EngineTypeCGo:
        select {
        case engine := <-ep.cgoEngines:
            return engine, nil
        default:
            return ep.factory.CreateCGo()
        }
    case EngineTypeWASM:
        select {
        case engine := <-ep.wasmEngines:
            return engine, nil
        default:
            return ep.factory.CreateWASM()
        }
    }
}

func (ep *EnginePool) Put(engine OCREngine) {
    switch engine.Type() {
    case EngineTypeCGo:
        select {
        case ep.cgoEngines <- engine.(*TesseractCGoEngine):
        default:
            engine.Close()
        }
    case EngineTypeWASM:
        select {
        case ep.wasmEngines <- engine.(*TesseractWASMEngine):
        default:
            engine.Close()
        }
    }
}
```

#### 4. Worker Pool (`internal/pool/worker_pool.go`)

```go
type WorkerPool struct {
    workers     int
    taskQueue   chan Task
    resultQueue chan Result
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

func (wp *WorkerPool) Submit(task Task) <-chan Result {
    resultChan := make(chan Result, 1)
    wp.taskQueue <- Task{
        Data:       task.Data,
        ResultChan: resultChan,
    }
    return resultChan
}
```

#### 5. 结果缓存 (`internal/cache/cache.go`)

```go
type Cache struct {
    store sync.Map
    ttl   time.Duration
}

type CacheKey struct {
    ImageHash string
    Language  string
    Options   string
}

func (c *Cache) Get(key CacheKey) (*ocr.Result, bool) {
    if value, ok := c.store.Load(key); ok {
        entry := value.(*CacheEntry)
        if time.Since(entry.Timestamp) < c.ttl {
            return entry.Result, true
        }
        c.store.Delete(key)
    }
    return nil, false
}
```

#### 6. MCP Tool Schema (`internal/tools/schemas.go`)

```go
// Input Schema
type OCRInput struct {
    ImagePath       string              `json:"image_path,omitempty"`
    ImageBase64     string              `json:"image_base64,omitempty"`
    ImageURL        string              `json:"image_url,omitempty"`
    Language        string              `json:"language"`
    EngineType      string              `json:"engine_type"`
    Preprocessing   PreprocessingConfig `json:"preprocessing"`
    PSM             int                 `json:"page_segmentation_mode"`
    ConfThreshold   float64             `json:"confidence_threshold"`
}

type PreprocessingConfig struct {
    Mode string `json:"mode"` // "auto", "fast", "standard", "high_quality", "custom"
}

// Output Schema
type OCROutput struct {
    Text                string          `json:"text"`
    Confidence          float64         `json:"confidence"`
    Words               []WordDetail    `json:"words"`
    PreprocessingApplied []string       `json:"preprocessing_applied"`
    EngineUsed          string          `json:"engine_used"`
    ProcessingTimeMs    int64           `json:"processing_time_ms"`
    ImageAnalysis       AnalysisResult  `json:"image_analysis,omitempty"`
}
```

## 关键实现要点

### 1. 双引擎支持策略

- 默认使用 CGo 引擎 (性能更好)
- 提供 WASM 引擎作为备选 (跨平台部署)
- 通过配置或 API 参数选择引擎类型
- 引擎切换对用户透明

### 2. 智能预处理算法

根据图像质量指标自动决定处理步骤:

```
图像分析 → 质量评估 → 动态选择处理步骤
├─ 亮度不足 → 增强对比度
├─ 噪点过多 → 去噪处理
├─ 倾斜明显 → 倾斜校正
└─ 质量良好 → 仅灰度化
```

**质量指标**:
- 亮度 (Brightness): 0-255
- 对比度 (Contrast): 标准差
- 清晰度 (Sharpness): Laplacian 方差
- 倾斜角度 (Skew Angle): -45° 到 45°
- 噪声水平 (Noise Level): 估算值

### 3. 性能优化措施

#### Worker Pool
- 固定大小的 worker 池 (默认 4-8 个)
- 任务队列缓冲 (默认 100)
- 优雅关闭和超时控制

#### 引擎资源池
- CGo 引擎池 (默认 4 个实例)
- WASM 引擎池 (默认 2 个实例)
- 空闲超时回收

#### 结果缓存
- LRU 缓存策略
- 基于图像哈希 + 参数的缓存键
- 可配置 TTL (默认 1 小时)
- 最大缓存条目限制

#### GoCV 内存管理
- 严格的 Mat 对象生命周期管理
- 使用 defer 确保资源释放
- 定期垃圾回收触发

### 4. 错误处理体系

```go
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

type OCRError struct {
    Code       ErrorCode              `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    StackTrace string                 `json:"stack_trace,omitempty"`
    Err        error                  `json:"-"`
}
```

### 5. 配置管理

```yaml
server:
  name: "mcp-ocr-server"
  version: "1.0.0"

ocr:
  default_engine: "cgo"  # "cgo" or "wasm"
  tessdata_path: "/usr/share/tesseract-ocr/4.00/tessdata"
  default_language: "eng"
  supported_languages: ["eng", "chi_sim", "chi_tra", "jpn"]
  default_psm: 3
  confidence_threshold: 60.0
  max_image_size_mb: 10

preprocessing:
  mode: "auto"  # "auto", "fast", "standard", "high_quality", "custom"
  quality_thresholds:
    low_brightness: 80
    low_contrast: 30
    high_noise: 15
    skew_angle_threshold: 1.0

performance:
  worker_pool_size: 8
  cgo_engine_pool_size: 4
  wasm_engine_pool_size: 2
  request_timeout: "30s"
  max_concurrent_requests: 20

cache:
  enable: true
  ttl: "1h"
  max_entries: 1000

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## 开发路线图

### 第一阶段: 基础框架 (2 周)

**目标**: 搭建项目基础架构

**任务**:
1. 初始化 Go 模块和项目结构
2. 配置管理和日志系统
3. 错误处理体系
4. 基础单元测试框架

**交付物**:
- 可运行的项目骨架
- 基础配置和日志功能
- README.md 和 ARCHITECTURE.md

### 第二阶段: CGo OCR 引擎 (2 周)

**目标**: 实现 Gosseract 引擎封装

**任务**:
1. Tesseract CGo 引擎实现
2. 多语言支持配置
3. 基础 OCR 功能
4. 单元测试 (覆盖率 > 80%)

**交付物**:
- `internal/ocr/tesseract_cgo.go`
- 单元测试和集成测试
- 语言包安装脚本

### 第三阶段: 图像预处理 (2 周)

**目标**: 实现智能预处理管道

**任务**:
1. GoCV 图像处理集成
2. 图像质量分析器
3. 四个预处理步骤实现
4. 智能预处理管道
5. 性能基准测试

**交付物**:
- `internal/preprocessing/` 模块
- 图像质量分析算法
- 预处理性能报告

### 第四阶段: MCP Server 集成 (2 周)

**目标**: 集成 MCP 协议

**任务**:
1. MCP Go SDK 集成
2. OCR Tool 实现
3. Input/Output Schema 定义
4. 错误处理和日志
5. 集成测试

**交付物**:
- 完整的 MCP Server
- API 文档
- 集成测试套件

### 第五阶段: 性能优化 (2 周)

**目标**: 实现完整性能优化

**任务**:
1. Worker Pool 实现
2. 引擎资源池
3. 结果缓存
4. 内存优化
5. 性能压测

**交付物**:
- 性能优化模块
- 压力测试报告
- 性能调优文档

### 第六阶段: WASM 引擎和收尾 (2 周)

**目标**: 添加 WASM 引擎并完善项目

**任务**:
1. Gogosseract WASM 引擎
2. 双引擎切换逻辑
3. Docker 镜像构建
4. 文档完善
5. 最终测试

**交付物**:
- WASM 引擎支持
- Dockerfile 和部署文档
- 完整项目文档
- v1.0.0 Release

## 技术难点和解决方案

### 1. CGo 跨平台编译

**难点**: CGo 依赖系统库,不同平台编译复杂

**解决方案**:
- 提供平台特定的构建脚本
- 使用构建标签分离平台代码
- Docker 多阶段构建
- 提供预编译二进制文件

### 2. GoCV 内存管理

**难点**: Mat 对象不受 GC 管理,容易内存泄漏

**解决方案**:
- 严格的资源释放规范
- 使用 defer 确保清理
- 内存监控和泄漏检测
- 定期 GC 触发

### 3. 智能预处理算法

**难点**: 如何准确判断图像质量并选择处理步骤

**解决方案**:
- 多维度图像质量评估
- 基于阈值的决策树
- 可配置的质量标准
- 预处理效果反馈循环

### 4. 并发性能优化

**难点**: OCR 是 CPU 密集型操作,需要高效并发

**解决方案**:
- Worker Pool 限制并发数
- 引擎资源池复用实例
- 任务队列缓冲
- 超时和取消机制

## 部署方案

### 依赖安装

**macOS**:
```bash
brew install tesseract opencv pkg-config
brew install tesseract-lang  # 多语言包
```

**Linux (Ubuntu/Debian)**:
```bash
sudo apt-get update
sudo apt-get install -y \
    libtesseract-dev \
    libleptonica-dev \
    libopencv-dev \
    pkg-config \
    tesseract-ocr-eng \
    tesseract-ocr-chi-sim \
    tesseract-ocr-chi-tra \
    tesseract-ocr-jpn
```

### Docker 部署

```dockerfile
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache \
    build-base \
    pkgconfig \
    opencv-dev \
    tesseract-ocr-dev \
    leptonica-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o mcp-ocr-server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache \
    opencv \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-chi_sim \
    tesseract-ocr-data-chi_tra \
    tesseract-ocr-data-jpn \
    leptonica

COPY --from=builder /app/mcp-ocr-server /usr/local/bin/
COPY configs/config.yaml /etc/mcp-ocr-server/config.yaml

ENTRYPOINT ["mcp-ocr-server"]
CMD ["--config", "/etc/mcp-ocr-server/config.yaml"]
```

### MCP 客户端配置

**Claude Desktop** (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "ocr-server": {
      "command": "/path/to/mcp-ocr-server",
      "args": ["--config", "/path/to/config.yaml"]
    }
  }
}
```

## 质量保证

### 测试策略

1. **单元测试**: 覆盖率 > 80%
2. **集成测试**: 核心流程全覆盖
3. **性能测试**: 吞吐量和延迟基准
4. **压力测试**: 并发负载测试

### 文档要求

- **README.md**: 项目介绍和快速开始
- **ARCHITECTURE.md**: 架构设计文档
- **API.md**: MCP Tool API 文档
- **DEPLOYMENT.md**: 部署和运维指南
- **DEVELOPMENT.md**: 开发者指南
- **代码注释**: 关键逻辑和复杂算法

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化
- 使用 `golangci-lint` 静态检查
- 错误处理必须完整
- 资源释放必须保证

## 关键文件清单

实施时需要重点关注的核心文件:

1. **cmd/server/main.go** - 服务启动入口
2. **internal/ocr/tesseract_cgo.go** - CGo 引擎实现
3. **internal/preprocessing/pipeline.go** - 智能预处理管道
4. **internal/tools/ocr_tool.go** - MCP Tool 实现
5. **internal/server/server.go** - MCP Server 封装
6. **internal/pool/worker_pool.go** - Worker Pool 实现
7. **internal/ocr/pool.go** - 引擎资源池

## 参考资源

### 官方文档
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Gosseract](https://github.com/otiai10/gosseract)
- [GoCV](https://gocv.io/)
- [Tesseract OCR](https://tesseract-ocr.github.io/)

### 技术文章
- [Building MCP Server in Go](https://navendu.me/posts/mcp-server-go/)
- [Preprocessing for Tesseract OCR](https://stackoverflow.com/questions/28935983/preprocessing-image-for-tesseract-ocr-with-opencv)
- [Image Quality Assessment for OCR](https://autbor.com/preprocessingocr/)

---

**计划制定日期**: 2025-11-29
**预计完成时间**: 12 周
**项目优先级**: 高
**技术风险**: 中等 (CGo 依赖管理和跨平台编译)