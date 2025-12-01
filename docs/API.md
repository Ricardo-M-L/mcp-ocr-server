# API 文档

MCP OCR Server 提供了一套完整的 OCR 工具，通过 Model Context Protocol (MCP) 进行调用。

## MCP Tools

### 1. ocr_recognize_text

从图像文件中识别文本。

**工具名称**: `ocr_recognize_text`

**参数**:

| 参数名 | 类型 | 必需 | 默认值 | 描述 |
|--------|------|------|--------|------|
| `image_path` | string | 是 | - | 图像文件的绝对路径 |
| `language` | string | 否 | `eng` | OCR 识别语言 |
| `preprocess` | boolean | 否 | `true` | 是否启用图像预处理 |
| `auto_mode` | boolean | 否 | `true` | 是否启用自动质量分析 |

**语言代码**:
- `eng` - 英文
- `chi_sim` - 简体中文
- `chi_tra` - 繁体中文
- `jpn` - 日文
- `eng+chi_sim` - 英文+简体中文 (可组合)

**请求示例**:

```json
{
  "tool": "ocr_recognize_text",
  "arguments": {
    "image_path": "/Users/ricardo/Documents/test-image.png",
    "language": "eng+chi_sim",
    "preprocess": true,
    "auto_mode": true
  }
}
```

**响应示例**:

```json
{
  "text": "This is the recognized text from the image.\n这是从图像中识别的文本。",
  "confidence": 95.5,
  "language": "eng+chi_sim",
  "duration": 1.234
}
```

**错误响应**:

```json
{
  "error": "file not found: /path/to/image.png",
  "code": "FILE_NOT_FOUND"
}
```

---

### 2. ocr_recognize_text_base64

从 Base64 编码的图像中识别文本。

**工具名称**: `ocr_recognize_text_base64`

**参数**:

| 参数名 | 类型 | 必需 | 默认值 | 描述 |
|--------|------|------|--------|------|
| `image_base64` | string | 是 | - | Base64 编码的图像数据 |
| `language` | string | 否 | `eng` | OCR 识别语言 |
| `preprocess` | boolean | 否 | `true` | 是否启用图像预处理 |
| `auto_mode` | boolean | 否 | `true` | 是否启用自动质量分析 |

**请求示例**:

```json
{
  "tool": "ocr_recognize_text_base64",
  "arguments": {
    "image_base64": "iVBORw0KGgoAAAANSUhEUgAAAAUA...",
    "language": "eng",
    "preprocess": true
  }
}
```

**响应格式**: 与 `ocr_recognize_text` 相同

---

### 3. ocr_batch_recognize

批量识别多个图像文件。

**工具名称**: `ocr_batch_recognize`

**参数**:

| 参数名 | 类型 | 必需 | 默认值 | 描述 |
|--------|------|------|--------|------|
| `image_paths` | array | 是 | - | 图像文件路径数组 |
| `language` | string | 否 | `eng` | OCR 识别语言 |
| `preprocess` | boolean | 否 | `true` | 是否启用图像预处理 |
| `auto_mode` | boolean | 否 | `true` | 是否启用自动质量分析 |

**请求示例**:

```json
{
  "tool": "ocr_batch_recognize",
  "arguments": {
    "image_paths": [
      "/path/to/image1.png",
      "/path/to/image2.jpg",
      "/path/to/image3.png"
    ],
    "language": "chi_sim",
    "preprocess": true
  }
}
```

**响应示例**:

```json
{
  "results": [
    {
      "path": "/path/to/image1.png",
      "text": "识别的文本内容 1",
      "confidence": 96.2,
      "language": "chi_sim",
      "duration": 1.1
    },
    {
      "path": "/path/to/image2.jpg",
      "text": "识别的文本内容 2",
      "confidence": 94.8,
      "language": "chi_sim",
      "duration": 0.9
    },
    {
      "path": "/path/to/image3.png",
      "error": "file not found"
    }
  ],
  "count": 3
}
```

**特性**:
- 并行处理所有图像
- 单个图像失败不影响其他图像
- 返回包含成功和失败的完整结果

---

### 4. ocr_get_supported_languages

获取支持的 OCR 语言列表。

**工具名称**: `ocr_get_supported_languages`

**参数**: 无

**请求示例**:

```json
{
  "tool": "ocr_get_supported_languages",
  "arguments": {}
}
```

**响应示例**:

```json
{
  "languages": [
    "eng",
    "chi_sim",
    "chi_tra",
    "jpn"
  ]
}
```

---

## 错误代码

| 错误代码 | 描述 |
|----------|------|
| `INVALID_INPUT` | 无效的输入参数 |
| `FILE_NOT_FOUND` | 图像文件不存在 |
| `UNSUPPORTED_FORMAT` | 不支持的图像格式 |
| `IMAGE_TOO_LARGE` | 图像文件超过大小限制 |
| `PREPROCESSING_FAILED` | 图像预处理失败 |
| `OCR_ENGINE_FAILED` | OCR 引擎执行失败 |
| `TIMEOUT` | 操作超时 |
| `INTERNAL_ERROR` | 内部服务器错误 |

---

## 图像预处理

当 `preprocess` 参数为 `true` 时，系统会对图像进行智能预处理以提高识别准确率。

### 自动模式 (auto_mode: true)

系统会自动分析图像质量并选择合适的预处理步骤:

1. **质量分析**:
   - 清晰度检测 (Laplacian 方差)
   - 对比度检测 (标准差)
   - 亮度检测 (平均值)

2. **自适应处理**:
   - 低清晰度 → 降噪
   - 低对比度 → 对比度增强
   - 亮度不足 → 亮度调整
   - 倾斜文本 → 倾斜校正

### 手动模式 (auto_mode: false)

使用配置文件中定义的固定预处理管道:

1. 灰度化
2. 降噪 (可选)
3. 二值化
4. 倾斜校正 (可选)

---

## 性能优化

### 缓存机制

- **缓存键**: 基于图像数据 + 参数的 SHA256 哈希
- **缓存时长**: 可配置 (默认 3600 秒)
- **缓存大小**: 可配置 (默认 100 条目)

重复识别相同图像会直接返回缓存结果，大幅提升性能。

### Worker Pool

- **并发处理**: 多个 Worker 并发处理 OCR 任务
- **队列管理**: 任务队列缓冲高峰请求
- **资源池**: Tesseract 客户端池复用

### 资源限制

可通过配置文件调整资源使用:

```yaml
ocr:
  max_image_size: 10485760  # 10MB
  timeout: 30               # 30秒

performance:
  worker_pool_size: 4       # 4 个 Worker
  queue_size: 100          # 队列大小 100
  cache_size: 100          # 缓存 100 条目
```

---

## 最佳实践

### 1. 图像质量

为获得最佳识别效果:

- **分辨率**: 300 DPI 或更高
- **格式**: PNG、JPEG (推荐 PNG)
- **对比度**: 文字与背景对比明显
- **清晰度**: 避免模糊、抖动
- **角度**: 文本保持水平

### 2. 语言选择

- 单语言文档使用单语言模式 (如 `eng`)
- 混合语言文档使用组合模式 (如 `eng+chi_sim`)
- 避免不必要的语言组合以提高准确率

### 3. 预处理选项

- **高质量扫描件**: `preprocess: false`
- **照片、截图**: `preprocess: true, auto_mode: true`
- **低质量图像**: `preprocess: true, auto_mode: true`

### 4. 批量处理

- 使用 `ocr_batch_recognize` 而非多次调用单图识别
- 并行处理可显著提升总体吞吐量
- 注意总体资源消耗

### 5. 错误处理

始终检查响应中的错误字段:

```javascript
if (result.error) {
  console.error(`OCR failed: ${result.error}`);
  // 处理错误
} else {
  console.log(`Recognized text: ${result.text}`);
}
```

---

## 使用示例

### Claude Desktop 集成

```
用户: 请识别这个文档的内容 /Users/ricardo/Documents/invoice.png

Claude: 我来使用 OCR 工具识别这个文档...

[调用 ocr_recognize_text]

识别成功！文档内容如下:

```
发票
发票号: INV-2024-001
日期: 2024-01-15
金额: ¥1,000.00
```

置信度: 97.2%
处理时间: 1.1秒
```

### 批量处理示例

```
用户: 批量识别 /Documents/receipts/ 目录下的所有收据

Claude: 我来批量识别这些收据...

[调用 ocr_batch_recognize]

识别完成！共处理 10 张收据:
- 成功: 9 张
- 失败: 1 张 (文件不存在)

所有识别结果:
1. receipt-001.png: "购物小票..."
2. receipt-002.jpg: "餐饮发票..."
...
```

---

## 版本历史

### v1.0.0 (2024-01-01)

初始版本，包含:
- 基础 OCR 功能
- 多语言支持
- 智能预处理
- 批量处理
- 结果缓存

---

## 技术支持

如有问题或建议，请:

1. 查看 [README.md](../README.md)
2. 查看 [快速入门指南](QUICKSTART.md)
3. 提交 [GitHub Issue](https://github.com/ricardo/mcp-ocr-server/issues)

---

**最后更新**: 2024-01-01
**版本**: 1.0.0