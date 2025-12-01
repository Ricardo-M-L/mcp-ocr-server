#!/bin/bash

set -e

echo "Installing dependencies for MCP OCR Server..."

# 检测操作系统
OS=$(uname -s)

install_macos() {
    echo "Detected macOS"

    # 检查 Homebrew
    if ! command -v brew &> /dev/null; then
        echo "Homebrew not found. Please install Homebrew first:"
        echo "/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi

    # 安装 Tesseract
    echo "Installing Tesseract..."
    brew install tesseract

    # 安装语言包
    echo "Installing language data..."
    brew install tesseract-lang

    # 安装 OpenCV
    echo "Installing OpenCV..."
    brew install opencv

    # 设置环境变量
    echo "Setting up environment variables..."
    export PKG_CONFIG_PATH="/usr/local/opt/opencv/lib/pkgconfig:$PKG_CONFIG_PATH"
    export DYLD_LIBRARY_PATH="/usr/local/opt/opencv/lib:$DYLD_LIBRARY_PATH"

    echo "✓ macOS dependencies installed"
}

install_ubuntu() {
    echo "Detected Ubuntu/Debian"

    # 更新包列表
    sudo apt-get update

    # 安装 Tesseract
    echo "Installing Tesseract..."
    sudo apt-get install -y tesseract-ocr

    # 安装语言包
    echo "Installing language data..."
    sudo apt-get install -y tesseract-ocr-eng \
        tesseract-ocr-chi-sim \
        tesseract-ocr-chi-tra \
        tesseract-ocr-jpn

    # 安装 OpenCV 依赖
    echo "Installing OpenCV dependencies..."
    sudo apt-get install -y \
        libopencv-dev \
        pkg-config

    echo "✓ Ubuntu/Debian dependencies installed"
}

install_centos() {
    echo "Detected CentOS/RHEL"

    # 安装 EPEL 仓库
    sudo yum install -y epel-release

    # 安装 Tesseract
    echo "Installing Tesseract..."
    sudo yum install -y tesseract tesseract-langpack-eng \
        tesseract-langpack-chi_sim \
        tesseract-langpack-chi_tra \
        tesseract-langpack-jpn

    # 安装 OpenCV
    echo "Installing OpenCV..."
    sudo yum install -y opencv-devel

    echo "✓ CentOS/RHEL dependencies installed"
}

# 根据操作系统安装依赖
case "$OS" in
    Darwin)
        install_macos
        ;;
    Linux)
        if [ -f /etc/debian_version ]; then
            install_ubuntu
        elif [ -f /etc/redhat-release ]; then
            install_centos
        else
            echo "Unsupported Linux distribution"
            exit 1
        fi
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# 验证安装
echo ""
echo "Verifying installation..."

if command -v tesseract &> /dev/null; then
    echo "✓ Tesseract version: $(tesseract --version | head -n1)"
else
    echo "✗ Tesseract not found"
    exit 1
fi

# 检查语言数据
TESSDATA_PREFIX=$(tesseract --print-parameters 2>/dev/null | grep tessdata | awk '{print $2}')
if [ -n "$TESSDATA_PREFIX" ]; then
    echo "✓ Tessdata path: $TESSDATA_PREFIX"
    echo "Available languages:"
    tesseract --list-langs 2>/dev/null | grep -E "(eng|chi_sim|chi_tra|jpn)" || true
else
    echo "⚠ Could not detect tessdata path"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Run 'make deps' to install Go dependencies"
echo "2. Run 'make build' to build the server"
echo "3. Run './bin/mcp-ocr-server' to start the server"