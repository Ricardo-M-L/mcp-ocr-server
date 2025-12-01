# 多阶段构建 Dockerfile

# 阶段 1: 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache \
    git \
    make \
    gcc \
    g++ \
    pkgconfig \
    tesseract-ocr \
    tesseract-ocr-dev \
    opencv \
    opencv-dev

WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 阶段 2: 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-chi_sim \
    tesseract-ocr-data-chi_tra \
    tesseract-ocr-data-jpn \
    opencv \
    ca-certificates

# 创建应用目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/bin/mcp-ocr-server /app/

# 复制配置文件
COPY configs /app/configs

# 设置环境变量
ENV TESSDATA_PREFIX=/usr/share/tessdata

# 暴露端口 (如果需要)
# EXPOSE 8080

# 运行应用
ENTRYPOINT ["/app/mcp-ocr-server"]
CMD ["-config", "/app/configs/config.yaml"]