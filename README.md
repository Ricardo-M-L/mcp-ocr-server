# MCP OCR Server

生产级 OCR MCP Server，基于 Tesseract OCR 和 GoCV，提供智能图像预处理和高性能文本识别服务。

## 特性

### 核心功能
- ✅ **多语言支持**: 英文、简体中文、繁体中文、日文
- ✅ **智能预处理**: 自动图像质量分析和自适应预处理管道
- ✅ **高性能**: Worker Pool + 资源池 + 结果缓存
- ✅ **MCP 集成**: 完整的 Model Context Protocol 支持
- ✅ **生产就绪**: 完善的错误处理、日志记录和配置管理

### 图像预处理
- 自动质量分析 (清晰度、对比度、亮度)
- 灰度化处理
- 降噪 (Fast Non-Local Means Denoising)
- 二值化 (Otsu / 自适应阈值)
- 倾斜校正 (基于霍夫变换)
- 对比度增强 (CLAHE)
- 亮度调整

### 性能优化
- Worker Pool 并发处理
- Tesseract 客户端池
- 基于 SHA256 的结果缓存
- 可配置的资源限制

## 系统要求

- Go 1.21+
- Tesseract OCR 4.0+
- OpenCV 4.5+
- macOS / Linux (Ubuntu, CentOS)

## 快速开始

### 1. 安装系统依赖

```bash
# macOS
./scripts/install-deps.sh

# 或手动安装
brew install tesseract tesseract-lang opencv
```

### 2. 安装 Go 依赖

```bash
make deps
```

### 3. 配置

编辑 `configs/config.yaml`:

```yaml
server:
  name: mcp-ocr-server
  version: 1.0.0

ocr:
  language: eng+chi_sim+chi_tra+jpn
  data_path: /usr/local/share/tessdata
  max_image_size: 10485760  # 10MB
  timeout: 30

preprocessing:
  enabled: true
  auto_mode: true  # 智能预处理

performance:
  worker_pool_size: 4
  cache_enabled: true
  cache_size: 100
```

### 4. 构建和运行

```bash
# 构建
make build

# 运行
make run

# 或直接运行
./bin/mcp-ocr-server -config configs/config.yaml
```

## MCP Tools

### 1. ocr_recognize_text

识别图像文件中的文本。

**参数:**
```json
{
  "image_path": "/path/to/image.png",
  "language": "eng",
  "preprocess": true,
  "auto_mode": true
}
```

**返回:**
```json
{
  "text": "识别的文本内容",
  "confidence": 95.5,
  "language": "eng",
  "duration": 1.23
}
```

### 2. ocr_recognize_text_base64

识别 Base64 编码的图像。

**参数:**
```json
{
  "image_base64": "iVBORw0KGgoAAAANSUhEUgA...",
  "language": "chi_sim",
  "preprocess": true,
  "auto_mode": true
}
```

### 3. ocr_batch_recognize

批量识别多个图像。

**参数:**
```json
{
  "image_paths": [
    "/path/to/image1.png",
    "/path/to/image2.jpg"
  ],
  "language": "eng+chi_sim",
  "preprocess": true
}
```

**返回:**
```json
{
  "results": [
    {
      "path": "/path/to/image1.png",
      "text": "...",
      "confidence": 95.5
    },
    {
      "path": "/path/to/image2.jpg",
      "text": "...",
      "confidence": 92.3
    }
  ],
  "count": 2
}
```

### 4. ocr_get_supported_languages

获取支持的语言列表。

**返回:**
```json
{
  "languages": ["eng", "chi_sim", "chi_tra", "jpn"]
}
```

## 使用示例

### Claude Desktop 集成

在 `claude_desktop_config.json` 中添加:

```json
{
  "mcpServers": {
    "ocr": {
      "command": "/path/to/mcp-ocr-server",
      "args": ["-config", "/path/to/config.yaml"]
    }
  }
}
```

### 示例对话

```
用户: 请识别这张图片中的文本 /path/to/document.png

Claude: 我来使用 OCR 工具识别这张图片...

[调用 ocr_recognize_text]

识别结果:
- 文本: "这是一份重要文档..."
- 置信度: 96.5%
- 语言: 简体中文
- 处理时间: 1.2秒
```

## 项目结构

```
mcp-ocr-server/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── ocr/
│   │   ├── engine.go            # OCR 引擎接口
│   │   └── tesseract.go         # Tesseract 实现
│   ├── preprocessing/
│   │   ├── analyzer.go          # 图像质量分析
│   │   └── preprocessor.go      # 图像预处理
│   ├── pool/
│   │   └── worker_pool.go       # Worker Pool
│   ├── cache/
│   │   └── cache.go             # 结果缓存
│   ├── tools/
│   │   ├── schemas.go           # MCP Tool Schema
│   │   └── handler.go           # Tool Handler
│   └── server/
│       └── server.go            # MCP Server
├── pkg/
│   ├── errors/
│   │   └── errors.go            # 错误处理
│   └── logger/
│       └── logger.go            # 日志封装
├── configs/
│   └── config.yaml              # 配置文件
├── scripts/
│   └── install-deps.sh          # 依赖安装脚本
├── Makefile                     # 构建管理
├── Dockerfile                   # Docker 支持
└── README.md                    # 项目文档
```

## 配置说明

### OCR 配置

```yaml
ocr:
  language: eng+chi_sim           # 语言组合
  data_path: /path/to/tessdata    # tessdata 路径
  page_seg_mode: 3                # 页面分割模式
  max_image_size: 10485760        # 最大图像大小
  timeout: 30                     # 超时时间
```

### 预处理配置

```yaml
preprocessing:
  enabled: true                   # 启用预处理
  auto_mode: true                 # 自动模式
  grayscale: true                 # 灰度化
  denoise: true                   # 降噪
  binarization: true              # 二值化
  deskew_correction: true         # 倾斜校正
  quality_thresholds:
    sharpness: 100.0              # 清晰度阈值
    contrast: 30.0                # 对比度阈值
    brightness: 50.0              # 亮度阈值
```

### 性能配置

```yaml
performance:
  worker_pool_size: 4             # Worker 数量
  queue_size: 100                 # 队列大小
  cache_enabled: true             # 启用缓存
  cache_size: 100                 # 缓存大小
  cache_ttl: 3600                 # 缓存 TTL
```

## Docker 部署

### 构建镜像

```bash
make docker-build
```

### 运行容器

```bash
make docker-run
```

或手动运行:

```bash
docker run --rm -it \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/test:/app/test \
  mcp-ocr-server:latest
```

## 开发指南

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 开发模式 (热重载)

```bash
make dev
```

## 性能调优

### 1. Worker Pool 大小

根据 CPU 核心数调整:
```yaml
performance:
  worker_pool_size: 8  # 建议为 CPU 核心数
```

### 2. 缓存策略

高频场景增加缓存:
```yaml
performance:
  cache_size: 500
  cache_ttl: 7200
```

### 3. 预处理优化

低质量图像启用完整预处理:
```yaml
preprocessing:
  auto_mode: true      # 自动分析
  denoise: true        # 降噪
  deskew_correction: true  # 倾斜校正
```

### 4. 资源限制

限制图像大小:
```yaml
ocr:
  max_image_size: 5242880  # 5MB
  timeout: 15              # 15秒
```

## 故障排查

### 1. Tesseract 找不到语言数据

```bash
# 检查 tessdata 路径
tesseract --print-parameters | grep tessdata

# 设置环境变量
export TESSDATA_PREFIX=/usr/local/share/tessdata
```

### 2. OpenCV 链接错误

```bash
# macOS
export PKG_CONFIG_PATH="/usr/local/opt/opencv/lib/pkgconfig"
export DYLD_LIBRARY_PATH="/usr/local/opt/opencv/lib"

# Ubuntu
sudo ldconfig
```

### 3. 内存使用过高

减小 Worker Pool 和缓存大小:
```yaml
performance:
  worker_pool_size: 2
  cache_size: 50
```

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

MIT License

## 致谢

- [Tesseract OCR](https://github.com/tesseract-ocr/tesseract)
- [GoCV](https://gocv.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Gosseract](https://github.com/otiai10/gosseract)

## 联系方式

- Issues: [GitHub Issues](https://github.com/ricardo/mcp-ocr-server/issues)
- Email: your-email@example.com

---

**注意**: 本项目处于活跃开发中，API 可能会发生变化。建议关注 Release 版本进行生产部署。