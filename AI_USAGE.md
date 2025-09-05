# YAML AI 总结功能使用指南

## 🤖 功能概述

YAML 项目现已集成 Gemini AI 总结功能，可以智能分析用户活动数据并生成总结报告。

## 🔧 技术实现

### 后端架构
- **AI 客户端**: `internal/ai/gemini.go` - 封装 Gemini API 调用
- **AI 服务**: `internal/ai/service.go` - 提供高级 AI 功能接口
- **数据存储**: SQLite 数据库存储 AI 总结结果
- **API 端点**: RESTful API 提供 AI 功能访问

### API 配置
- **API Key**: `sk-JIyFjsX1HIuusXty13315a05E29440D88369B8797159E3A4`
- **Base URL**: `https://aihubmix.com/gemini`
- **模型**: `gemini-2.5-flash`

## 📡 API 端点

### 1. 生成活动总结
```bash
POST /api/v1/ai/summary/activity?limit=20
```

**功能**: 分析用户最近的活动记录，生成智能总结

**参数**:
- `limit` (可选): 分析的活动记录数量，默认 20

**响应示例**:
```json
{
  "id": 1,
  "type": "activity",
  "summary": "根据您最近的活动分析...",
  "data_count": 15,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 2. 生成键盘输入总结
```bash
POST /api/v1/ai/summary/keyboard?limit=15
```

**功能**: 分析用户键盘输入模式，生成使用习惯总结

**参数**:
- `limit` (可选): 分析的输入记录数量，默认 15

**响应示例**:
```json
{
  "id": 2,
  "type": "keyboard",
  "summary": "您的输入习惯分析显示...",
  "data_count": 12,
  "created_at": "2024-01-15T10:35:00Z"
}
```

### 3. 获取历史总结
```bash
GET /api/v1/ai/summaries?limit=10
```

**功能**: 获取之前生成的 AI 总结历史记录

**参数**:
- `limit` (可选): 返回的总结数量，默认 10

**响应示例**:
```json
{
  "summaries": [
    {
      "id": 2,
      "type": "keyboard",
      "summary": "您的输入习惯分析...",
      "data_count": 12,
      "created_at": "2024-01-15T10:35:00Z"
    },
    {
      "id": 1,
      "type": "activity",
      "summary": "根据您最近的活动分析...",
      "data_count": 15,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 2
}
```

## 🚀 使用示例

### 启动服务器
```bash
cd backend
go run cmd/server/main.go
```

### 启动监控（生成测试数据）
```bash
curl -X POST http://localhost:8080/api/v1/monitor/start
```

### 生成活动总结
```bash
curl -X POST "http://localhost:8080/api/v1/ai/summary/activity?limit=10"
```

### 生成键盘总结
```bash
curl -X POST "http://localhost:8080/api/v1/ai/summary/keyboard?limit=5"
```

### 查看总结历史
```bash
curl "http://localhost:8080/api/v1/ai/summaries?limit=5"
```

## 🔒 隐私保护

- **本地处理**: 所有数据在本地处理，仅总结结果发送给 AI
- **数据脱敏**: 键盘输入分析时不直接暴露具体内容
- **用户控制**: 用户完全控制何时生成总结
- **数据限制**: 每次分析的数据量有限制，避免过度暴露

## 📊 AI 分析维度

### 活动分析
1. 主要使用的应用程序
2. 活动时间分布
3. 工作效率评估
4. 建议和改进点

### 键盘输入分析
1. 输入内容的类型和特征
2. 使用频率最高的应用
3. 输入模式和习惯
4. 可能的工作内容推测（保护隐私）

## 🛠️ 开发说明

### 添加新的 AI 功能
1. 在 `internal/ai/gemini.go` 中添加新的 AI 调用方法
2. 在 `internal/ai/service.go` 中添加业务逻辑
3. 在 `internal/api/handlers.go` 中添加 API 处理器
4. 在 `internal/api/routes.go` 中注册新路由

### 自定义 AI 模型
修改 `internal/ai/gemini.go` 中的模型配置：
```go
apiURL := fmt.Sprintf("%s/v1beta/models/your-model:generateContent", g.BaseURL)
```

### 调整 AI 参数
在 `generateContent` 方法中修改生成配置：
```go
Config: Config{
    Temperature:     0.7,  // 创造性 (0-1)
    MaxOutputTokens: 6717, // 最大输出长度
    TopP:            0.8,  // 核采样
    TopK:            40,   // Top-K 采样
}
```

## 🔧 故障排除

### 常见问题
1. **API 调用失败**: 检查网络连接和 API Key
2. **权限错误**: 确保已授予辅助功能权限
3. **数据为空**: 确保监控已启动并有数据生成

### 调试方法
- 查看服务器日志输出
- 使用 `curl` 命令测试 API
- 检查数据库中的数据：`sqlite3 ~/.yaml/yaml.db`

---

**注意**: 请确保 API Key 的安全性，不要在公共代码库中暴露真实的 API Key。