# MCP OCR Server - 安装指南

本文档提供详细的安装说明,帮助您在不同操作系统上安装和配置 MCP OCR Server。

## 目录

- [系统要求](#系统要求)
- [快速安装](#快速安装)
- [详细安装步骤](#详细安装步骤)
  - [macOS 安装](#macos-安装)
  - [Ubuntu/Debian 安装](#ubuntudebian-安装)
  - [CentOS/RHEL 安装](#centosrhel-安装)
- [验证安装](#验证安装)
- [常见问题](#常见问题)

## 系统要求

### 最低配置
- **操作系统**: macOS 10.15+, Ubuntu 18.04+, CentOS 7+
- **CPU**: 2 核心
- **内存**: 4 GB RAM
- **硬盘**: 2 GB 可用空间
- **Go**: 1.21 或更高版本

### 推荐配置
- **CPU**: 4+ 核心
- **内存**: 8+ GB RAM
- **硬盘**: 5+ GB 可用空间 (用于语言数据和缓存)

### 必需的依赖
- **Tesseract OCR** 4.1.0+
- **OpenCV** 4.5.0+
- **pkg-config**
- **语言数据包**: eng, chi_sim, chi_tra, jpn

## 快速安装

如果您的系统已经安装了 Go,可以使用自动化安装脚本:

```bash
# 克隆或下载项目
cd /path/to/mcp-ocr-server

# 运行安装脚本 (会自动检测操作系统并安装依赖)
chmod +x scripts/install-deps.sh
./scripts/install-deps.sh

# 安装 Go 依赖
make deps

# 构建项目
make build

# 运行服务
./bin/mcp-ocr-server
```

## 详细安装步骤

### macOS 安装

#### 1. 安装 Homebrew (如果尚未安装)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. 安装依赖

```bash
# 安装 Tesseract OCR
brew install tesseract

# 安装语言数据包
brew install tesseract-lang

# 安装 OpenCV
brew install opencv

# 安装 pkg-config
brew install pkg-config
```

#### 3. 设置环境变量

将以下内容添加到 `~/.zshrc` 或 `~/.bash_profile`:

```bash
# OpenCV 配置
export PKG_CONFIG_PATH="/usr/local/opt/opencv/lib/pkgconfig:$PKG_CONFIG_PATH"
export DYLD_LIBRARY_PATH="/usr/local/opt/opencv/lib:$DYLD_LIBRARY_PATH"

# Tesseract 数据路径
export TESSDATA_PREFIX="/usr/local/share/tessdata"
```

重新加载配置:

```bash
source ~/.zshrc  # 或 source ~/.bash_profile
```

#### 4. 验证 Tesseract 安装

```bash
tesseract --version
tesseract --list-langs
```

应该看到包含 eng, chi_sim, chi_tra, jpn 的语言列表。

#### 5. 安装 Go 依赖和构建

```bash
cd /path/to/mcp-ocr-server

# 安装 Go 依赖
go mod download
go mod tidy

# 构建项目
go build -o bin/mcp-ocr-server ./cmd/server

# 或使用 Makefile
make deps
make build
```

### Ubuntu/Debian 安装

#### 1. 更新包列表

```bash
sudo apt-get update
```

#### 2. 安装 Tesseract 和语言包

```bash
# 安装 Tesseract OCR
sudo apt-get install -y tesseract-ocr

# 安装语言数据包
sudo apt-get install -y \
    tesseract-ocr-eng \
    tesseract-ocr-chi-sim \
    tesseract-ocr-chi-tra \
    tesseract-ocr-jpn
```

#### 3. 安装 OpenCV 依赖

```bash
sudo apt-get install -y \
    libopencv-dev \
    pkg-config
```

#### 4. 设置环境变量

将以下内容添加到 `~/.bashrc`:

```bash
# Tesseract 数据路径
export TESSDATA_PREFIX="/usr/share/tesseract-ocr/4.00/tessdata"

# OpenCV 配置
export PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig:$PKG_CONFIG_PATH"
```

重新加载配置:

```bash
source ~/.bashrc
```

#### 5. 验证和构建

```bash
# 验证 Tesseract
tesseract --version
tesseract --list-langs

# 构建项目
cd /path/to/mcp-ocr-server
make deps
make build
```

### CentOS/RHEL 安装

#### 1. 安装 EPEL 仓库

```bash
sudo yum install -y epel-release
```

#### 2. 安装 Tesseract 和语言包

```bash
sudo yum install -y \
    tesseract \
    tesseract-langpack-eng \
    tesseract-langpack-chi_sim \
    tesseract-langpack-chi_tra \
    tesseract-langpack-jpn
```

#### 3. 安装 OpenCV

```bash
sudo yum install -y opencv-devel pkg-config
```

#### 4. 设置环境变量

将以下内容添加到 `~/.bashrc`:

```bash
export TESSDATA_PREFIX="/usr/share/tesseract/tessdata"
export PKG_CONFIG_PATH="/usr/lib64/pkgconfig:$PKG_CONFIG_PATH"
```

重新加载配置:

```bash
source ~/.bashrc
```

#### 5. 验证和构建

```bash
tesseract --version
cd /path/to/mcp-ocr-server
make deps
make build
```

## 验证安装

### 1. 检查系统依赖

```bash
# 检查 Tesseract
which tesseract
tesseract --version

# 检查 OpenCV
pkg-config --modversion opencv4

# 检查语言数据
tesseract --list-langs
```

### 2. 测试 Go 构建

```bash
cd /path/to/mcp-ocr-server

# 运行测试
make test

# 构建项目
make build

# 运行服务
./bin/mcp-ocr-server --help
```

### 3. 测试 OCR 功能

创建测试脚本 `test_ocr.sh`:

```bash
#!/bin/bash

# 下载测试图片
curl -o test.png "https://tesseract-ocr.github.io/docs/img/testocr.png"

# 测试 Tesseract 命令行
tesseract test.png stdout

# 测试 MCP OCR Server (需要先启动服务)
# ./bin/mcp-ocr-server &
# sleep 2
# # 通过 MCP 协议测试...
```

## 配置

### 基本配置

复制示例配置文件:

```bash
cp configs/config.yaml configs/config.local.yaml
```

编辑 `configs/config.local.yaml`:

```yaml
server:
  name: mcp-ocr-server
  version: 1.0.0

ocr:
  engine: tesseract
  language: eng+chi_sim+chi_tra+jpn
  data_path: /usr/local/share/tessdata  # 根据您的系统调整

preprocessing:
  enabled: true
  auto_mode: true

performance:
  worker_pool_size: 4
  cache_enabled: true
  cache_size: 100

logger:
  level: info
  format: console
  output_path: stdout
```

### 运行服务

```bash
# 使用默认配置
./bin/mcp-ocr-server

# 使用自定义配置
./bin/mcp-ocr-server --config configs/config.local.yaml

# 使用环境变量
export TESSDATA_PREFIX=/path/to/tessdata
export OCR_LOG_LEVEL=debug
./bin/mcp-ocr-server
```

## 常见问题

### 1. `leptonica/allheaders.h' file not found`

**问题**: Tesseract 依赖的 Leptonica 库未安装。

**解决方案**:

```bash
# macOS
brew install leptonica

# Ubuntu/Debian
sudo apt-get install libleptonica-dev

# CentOS/RHEL
sudo yum install leptonica-devel
```

### 2. `pkg-config: executable file not found`

**问题**: pkg-config 未安装或不在 PATH 中。

**解决方案**:

```bash
# macOS
brew install pkg-config

# Ubuntu/Debian
sudo apt-get install pkg-config

# CentOS/RHEL
sudo yum install pkgconfig
```

### 3. `Failed to init API, possibly an invalid tessdata path`

**问题**: Tesseract 无法找到语言数据文件。

**解决方案**:

1. 确认 tessdata 路径:
```bash
tesseract --print-parameters 2>/dev/null | grep tessdata
```

2. 设置环境变量:
```bash
export TESSDATA_PREFIX=/path/to/tessdata
```

3. 在配置文件中指定路径:
```yaml
ocr:
  data_path: /usr/local/share/tessdata
```

### 4. `error while loading shared libraries: libopencv`

**问题**: OpenCV 库无法加载。

**解决方案**:

```bash
# macOS
export DYLD_LIBRARY_PATH="/usr/local/opt/opencv/lib:$DYLD_LIBRARY_PATH"

# Linux
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

# 或者运行 ldconfig (需要 root 权限)
sudo ldconfig
```

### 5. 语言包缺失

**问题**: OCR 无法识别特定语言。

**解决方案**:

```bash
# 检查已安装的语言
tesseract --list-langs

# macOS - 安装所有语言包
brew install tesseract-lang

# Ubuntu/Debian - 安装特定语言
sudo apt-get install tesseract-ocr-chi-sim

# CentOS/RHEL
sudo yum install tesseract-langpack-chi_sim
```

### 6. 构建时内存不足

**问题**: 编译过程中内存耗尽。

**解决方案**:

```bash
# 限制并发构建
go build -p 1 -o bin/mcp-ocr-server ./cmd/server

# 或使用 Makefile
make build BUILD_FLAGS="-p 1"
```

### 7. CGo 编译错误

**问题**: CGo 相关的编译错误。

**解决方案**:

```bash
# 确保 CGO 已启用
export CGO_ENABLED=1

# 设置 C 编译器
export CC=gcc  # 或 clang

# 查看详细错误信息
go build -x ./cmd/server
```

## 下一步

安装完成后,请参阅:

- [快速开始指南](QUICKSTART.md) - 快速上手使用
- [API 文档](API.md) - 详细的 API 参考
- [架构概览](OVERVIEW.md) - 了解系统架构
- [配置说明](../configs/README.md) - 配置选项详解

## 获取帮助

如果您遇到问题:

1. 查看 [常见问题](#常见问题) 部分
2. 检查 GitHub Issues
3. 查看项目文档
4. 联系维护者

## 许可证

本项目采用 MIT 许可证。