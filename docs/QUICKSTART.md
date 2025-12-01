# 快速入门指南

本指南将帮助你快速设置和运行 MCP OCR Server。

## 前置要求

1. Go 1.21 或更高版本
2. Tesseract OCR 4.0+
3. OpenCV 4.5+

## 步骤 1: 安装系统依赖

### macOS

```bash
# 使用自动安装脚本
./scripts/install-deps.sh

# 或手动安装
brew install tesseract tesseract-lang opencv
```

验证安装:
```bash
tesseract --version
tesseract --list-langs | grep -E "(eng|chi_sim|chi_tra|jpn)"
```

### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y \
  tesseract-ocr \
  tesseract-ocr-eng \
  tesseract-ocr-chi-sim \
  tesseract-ocr-chi-tra \
  tesseract-ocr-jpn \
  libopencv-dev
```

### CentOS/RHEL

```bash
sudo yum install -y epel-release
sudo yum install -y \
  tesseract \
  tesseract-langpack-eng \
  tesseract-langpack-chi_sim \
  tesseract-langpack-chi_tra \
  tesseract-langpack-jpn \
  opencv-devel
```

## 步骤 2: 克隆和构建

```bash
# 克隆仓库
cd /path/to/your/workspace
git clone https://github.com/ricardo/mcp-ocr-server.git
cd mcp-ocr-server

# 安装 Go 依赖
make deps

# 构建
make build
```

## 步骤 3: 配置

编辑 `configs/config.yaml`:

```yaml
ocr:
  language: eng+chi_sim  # 根据需要修改语言
  data_path: /usr/local/share/tessdata  # 确保路径正确

preprocessing:
  enabled: true
  auto_mode: true  # 启用智能预处理

performance:
  worker_pool_size: 4  # 根据 CPU 核心数调整

logger:
  level: info
  format: console
```

## 步骤 4: 运行

### 方式 1: 直接运行

```bash
./bin/mcp-ocr-server -config configs/config.yaml
```

### 方式 2: 使用 Make

```bash
make run
```

### 方式 3: Docker

```bash
make docker-build
make docker-run
```

## 步骤 5: 集成到 Claude Desktop

编辑 Claude Desktop 配置文件:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "ocr": {
      "command": "/path/to/mcp-ocr-server/bin/mcp-ocr-server",
      "args": ["-config", "/path/to/mcp-ocr-server/configs/config.yaml"]
    }
  }
}
```

重启 Claude Desktop。

## 步骤 6: 测试

### 简单测试

```bash
# 使用测试程序
go run test/simple/main.go /path/to/test-image.png
```

### Claude Desktop 测试

在 Claude Desktop 中输入:

```
请使用 OCR 工具识别这张图片中的文本: /path/to/image.png
```

## 验证安装

检查所有组件:

```bash
# 检查 Tesseract
tesseract --version

# 检查语言数据
tesseract --list-langs

# 检查 Go 环境
go version

# 检查构建
./bin/mcp-ocr-server -version
```

## 常见问题

### 1. "tessdata not found" 错误

设置环境变量:
```bash
export TESSDATA_PREFIX=/usr/local/share/tessdata
```

或在配置文件中指定:
```yaml
ocr:
  data_path: /usr/local/share/tessdata
```

### 2. OpenCV 链接错误

macOS:
```bash
export PKG_CONFIG_PATH="/usr/local/opt/opencv/lib/pkgconfig"
export DYLD_LIBRARY_PATH="/usr/local/opt/opencv/lib"
```

Ubuntu:
```bash
sudo ldconfig
```

### 3. 语言包缺失

安装特定语言包:
```bash
# macOS
brew install tesseract-lang

# Ubuntu
sudo apt-get install tesseract-ocr-chi-sim

# 验证
tesseract --list-langs
```

### 4. 权限错误

确保二进制文件有执行权限:
```bash
chmod +x bin/mcp-ocr-server
chmod +x scripts/install-deps.sh
```

## 下一步

- 阅读完整的 [README.md](../README.md)
- 查看 [配置说明](../configs/config.yaml)
- 探索 [API 文档](../docs/API.md)
- 运行测试: `make test`

## 获取帮助

- GitHub Issues: https://github.com/ricardo/mcp-ocr-server/issues
- 文档: https://github.com/ricardo/mcp-ocr-server/docs

---

祝你使用愉快！ 🎉