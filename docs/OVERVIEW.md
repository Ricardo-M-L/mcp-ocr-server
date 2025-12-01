# MCP OCR Server - 项目概览

## 项目信息

- **项目名称**: MCP OCR Server
- **版本**: 1.0.0
- **描述**: 生产级 OCR MCP Server，基于 Tesseract OCR 和 GoCV
- **开发语言**: Go 1.21+
- **协议**: Model Context Protocol (MCP)

## 核心特性

### 1. OCR 引擎
- ✅ Tesseract OCR 4.0+ 集成
- ✅ 客户端池复用机制
- ✅ 多语言支持 (英文、简繁中文、日文)
- ✅ 可配置的页面分割模式
- ✅ 置信度评估

### 2. 图像预处理
- ✅ 智能质量分析器
  - 清晰度检测 (Laplacian 方差)
  - 对比度检测 (标准差)
  - 亮度检测 (平均值)
- ✅ 自适应预处理管道
  - 灰度化
  - 降噪 (Fast NLMeans)
  - 二值化 (Otsu/自适应)
  - 倾斜校正 (霍夫变换)
  - 对比度增强 (CLAHE)
  - 亮度调整

### 3. 性能优化
- ✅ Worker Pool 并发处理
- ✅ Tesseract 客户端池
- ✅ 基于 SHA256 的结果缓存
- ✅ 可配置的资源限制
- ✅ 批量处理支持

### 4. MCP 集成
- ✅ 完整的 MCP SDK 集成
- ✅ 4 个核心工具
- ✅ 结构化的请求/响应
- ✅ 完善的错误处理

## 项目架构

```
┌─────────────────────────────────────────────────────┐
│                   MCP Client                        │
│              (Claude Desktop)                       │
└──────────────────┬──────────────────────────────────┘
                   │ MCP Protocol
                   ▼
┌─────────────────────────────────────────────────────┐
│                 MCP Server                          │
│  ┌──────────────────────────────────────────────┐  │
│  │           Tool Handler                       │  │
│  │  - ocr_recognize_text                       │  │
│  │  - ocr_recognize_text_base64                │  │
│  │  - ocr_batch_recognize                      │  │
│  │  - ocr_get_supported_languages              │  │
│  └─────────┬────────────────────────────────────┘  │
└────────────┼────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│              Processing Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │
│  │ Worker Pool  │  │    Cache     │  │  Config  │ │
│  └──────────────┘  └──────────────┘  └──────────┘ │
└─────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│              Core Services                          │
│  ┌──────────────────┐  ┌──────────────────────┐   │
│  │  Preprocessor    │  │   OCR Engine         │   │
│  │  - Analyzer      │  │   - Tesseract        │   │
│  │  - Filters       │  │   - Client Pool      │   │
│  └──────────────────┘  └──────────────────────┘   │
└─────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│          External Dependencies                      │
│  ┌──────────────────┐  ┌──────────────────────┐   │
│  │  GoCV/OpenCV     │  │  Tesseract OCR       │   │
│  └──────────────────┘  └──────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

## 目录结构

```
mcp-ocr-server/
├── cmd/                        # 应用程序入口
│   └── server/
│       └── main.go            # 服务主程序
│
├── internal/                   # 内部包
│   ├── cache/                 # 结果缓存
│   │   ├── cache.go
│   │   └── cache_test.go
│   │
│   ├── config/                # 配置管理
│   │   └── config.go
│   │
│   ├── ocr/                   # OCR 引擎
│   │   ├── engine.go          # 引擎接口
│   │   ├── tesseract.go       # Tesseract 实现
│   │   └── tesseract_test.go
│   │
│   ├── preprocessing/         # 图像预处理
│   │   ├── analyzer.go        # 质量分析器
│   │   └── preprocessor.go    # 预处理器
│   │
│   ├── pool/                  # Worker Pool
│   │   ├── worker_pool.go
│   │   └── worker_pool_test.go
│   │
│   ├── server/                # MCP Server
│   │   └── server.go
│   │
│   └── tools/                 # MCP Tools
│       ├── schemas.go         # Tool Schema 定义
│       └── handler.go         # Tool 处理器
│
├── pkg/                       # 公共包
│   ├── errors/               # 错误处理
│   │   └── errors.go
│   └── logger/               # 日志封装
│       └── logger.go
│
├── configs/                   # 配置文件
│   ├── config.yaml           # 生产配置
│   └── config.dev.yaml       # 开发配置
│
├── scripts/                   # 脚本
│   └── install-deps.sh       # 依赖安装
│
├── test/                      # 测试
│   └── simple/
│       └── main.go           # 简单测试程序
│
├── docs/                      # 文档
│   ├── API.md                # API 文档
│   ├── QUICKSTART.md         # 快速入门
│   └── PLAN.md               # 项目计划
│
├── .air.toml                  # Air 热重载配置
├── .golangci.yml              # Linter 配置
├── .gitignore                 # Git 忽略文件
├── Dockerfile                 # Docker 镜像
├── Makefile                   # 构建管理
├── README.md                  # 项目说明
├── go.mod                     # Go 模块
└── go.sum                     # 依赖锁定
```

## 技术栈

### 核心依赖

| 包名 | 版本 | 用途 |
|------|------|------|
| github.com/modelcontextprotocol/go-sdk | v0.1.0 | MCP 协议实现 |
| github.com/otiai10/gosseract/v2 | v2.4.1 | Tesseract OCR Go 绑定 |
| gocv.io/x/gocv | v0.35.0 | OpenCV Go 绑定 |
| go.uber.org/zap | v1.26.0 | 高性能日志库 |
| gopkg.in/yaml.v3 | v3.0.1 | YAML 配置解析 |

### 外部依赖

- **Tesseract OCR** 4.0+: 开源 OCR 引擎
- **OpenCV** 4.5+: 计算机视觉库
- **tessdata**: Tesseract 语言数据包

## 工作流程

### 1. 单图像识别流程

```
1. 接收 MCP 请求
   ├── 验证参数
   └── 提取图像路径/数据

2. 读取图像
   ├── 检查文件大小
   ├── 检查缓存
   └── 读取图像数据

3. 图像预处理 (如果启用)
   ├── 质量分析
   │   ├── 清晰度检测
   │   ├── 对比度检测
   │   └── 亮度检测
   └── 自适应处理
       ├── 灰度化
       ├── 降噪
       ├── 二值化
       └── 倾斜校正

4. OCR 识别
   ├── 从池中获取客户端
   ├── 配置参数
   ├── 执行识别
   └── 归还客户端

5. 结果处理
   ├── 提取文本
   ├── 计算置信度
   ├── 缓存结果
   └── 返回响应
```

### 2. 批量处理流程

```
1. 接收批量请求
   └── 解析图像路径列表

2. 并发处理
   ├── 为每个图像创建 Goroutine
   ├── Worker Pool 调度
   └── 收集所有结果

3. 汇总返回
   ├── 整合成功结果
   ├── 记录失败信息
   └── 返回批量响应
```

## 配置系统

### 配置层级

```
1. 默认配置 (config.GetDefault())
   ↓
2. 配置文件 (configs/config.yaml)
   ↓
3. 环境变量 (可选)
   ↓
4. 命令行参数 (-config flag)
```

### 主要配置项

```yaml
# OCR 引擎配置
ocr:
  language: eng+chi_sim+chi_tra+jpn
  max_image_size: 10485760
  timeout: 30

# 预处理配置
preprocessing:
  enabled: true
  auto_mode: true
  quality_thresholds:
    sharpness: 100.0
    contrast: 30.0
    brightness: 50.0

# 性能配置
performance:
  worker_pool_size: 4
  cache_enabled: true
  cache_size: 100

# 日志配置
logger:
  level: info
  format: console
```

## 性能指标

### 预期性能 (参考值)

| 场景 | 图像大小 | 处理时间 | 并发数 |
|------|----------|----------|--------|
| 简单文本 | 1MB | 0.5-1.0s | 4 |
| 复杂文档 | 3MB | 1.5-2.5s | 4 |
| 低质量图像 | 2MB | 2.0-3.0s | 4 |
| 批量处理 (10张) | 10MB | 3.0-5.0s | 4 |

### 优化策略

1. **缓存命中**: 相同图像 < 10ms
2. **并发处理**: Worker Pool 提升 3-4x
3. **预处理优化**: 自动模式减少不必要处理
4. **资源池**: 客户端复用减少初始化开销

## 部署方式

### 1. 本地开发

```bash
make build && make run
```

### 2. Docker 部署

```bash
make docker-build
make docker-run
```

### 3. Claude Desktop 集成

编辑配置文件:
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

### 4. 生产部署

```bash
# 构建
make build

# 安装为系统服务 (systemd)
sudo cp mcp-ocr-server.service /etc/systemd/system/
sudo systemctl enable mcp-ocr-server
sudo systemctl start mcp-ocr-server
```

## 开发指南

### 运行测试

```bash
# 单元测试
make test

# 覆盖率
make coverage

# 集成测试
go run test/simple/main.go /path/to/image.png
```

### 代码质量

```bash
# 格式化
make fmt

# Lint 检查
make lint
```

### 热重载开发

```bash
make dev  # 需要 air
```

## 故障排查

### 常见问题

1. **Tesseract 找不到语言数据**
   - 设置 `TESSDATA_PREFIX` 环境变量
   - 或在配置文件中指定 `data_path`

2. **OpenCV 链接错误**
   - 设置 `PKG_CONFIG_PATH`
   - 运行 `sudo ldconfig`

3. **内存使用过高**
   - 减少 `worker_pool_size`
   - 减少 `cache_size`
   - 降低 `max_image_size`

## 路线图

### v1.1.0 (计划中)
- [ ] PDF 文档 OCR 支持
- [ ] 表格识别功能
- [ ] 手写体识别
- [ ] GPU 加速支持

### v1.2.0 (计划中)
- [ ] REST API 接口
- [ ] WebSocket 实时流式传输
- [ ] 更多语言支持
- [ ] 自定义模型训练

### v2.0.0 (未来)
- [ ] 分布式部署支持
- [ ] 横向扩展能力
- [ ] 监控和指标系统
- [ ] Web 管理界面

## 贡献者

- Ricardo - 项目创建者和维护者

## 许可证

MIT License - 详见 LICENSE 文件

## 致谢

感谢以下开源项目:
- Tesseract OCR Team
- GoCV Contributors
- MCP SDK Team
- Gosseract Author

---

**最后更新**: 2024-01-01
**当前版本**: 1.0.0
**状态**: ✅ 生产就绪