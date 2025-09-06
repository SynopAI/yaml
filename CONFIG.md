# YAML 项目配置说明

## 概述

本项目现在支持通过配置文件来管理各种设置，包括API密钥、数据库路径、服务器端口等。这样可以方便地修改配置而无需重新编译代码。

## 配置文件位置

### 后端配置
- **文件路径**: `backend/config/config.yaml`
- **环境变量**: 可通过 `CONFIG_PATH` 环境变量指定自定义配置文件路径

### 前端配置
- **Web前端**: `frontend/web/config.js`
- **Swift应用**: 通过 `Info.plist` 中的 `APIBaseURL` 键配置

## 配置项说明

### 后端配置 (config.yaml)

#### 服务器配置
```yaml
server:
  port: "8080"        # 服务器端口
  host: "localhost"    # 服务器主机
```

#### 数据库配置
```yaml
database:
  type: "sqlite"       # 数据库类型
  filename: "yaml.db"  # 数据库文件名
  data_dir: ".yaml"    # 数据目录（相对于用户主目录）
  retention_days: 30   # 数据保留天数
```

#### AI服务配置
```yaml
ai:
  gemini:
    api_key: "your-api-key"              # Gemini API密钥
    base_url: "https://aihubmix.com/gemini" # API基础URL
    timeout_seconds: 120                     # 请求超时时间（秒）
    generation:
      temperature: 0.7        # 创造性参数 (0-1)
      max_output_tokens: 6717 # 最大输出token数
      top_p: 0.8             # 核采样参数
      top_k: 40              # Top-K采样参数
```

#### 监控配置
```yaml
monitor:
  collection_interval: 5    # 数据采集间隔（秒）
  keyboard_buffer_size: 1000 # 键盘输入缓冲区大小
  app_switch_interval: 500   # 应用切换检测间隔（毫秒）
```

#### API配置
```yaml
api:
  cors_origins:              # CORS允许的源
    - "http://localhost:3000"
    - "http://127.0.0.1:3000"
  default_limit: 50          # 默认查询限制
  max_limit: 1000           # 最大查询限制
```

#### 日志配置
```yaml
logging:
  level: "info"              # 日志级别 (debug, info, warn, error)
  file_output: false         # 是否输出到文件
  file_path: "logs/yaml.log" # 日志文件路径
```

### 前端配置 (config.js)

#### API配置
```javascript
API: {
    BASE_URL: 'http://localhost:8080/api/v1', // API基础地址
    TIMEOUT: 30000,                           // 请求超时时间（毫秒）
}
```

#### 自动刷新配置
```javascript
AUTO_REFRESH: {
    ENABLED: true,  // 是否启用自动刷新
    INTERVAL: 5000, // 刷新间隔（毫秒）
}
```

#### 数据显示配置
```javascript
DISPLAY: {
    DEFAULT_LIMIT: 20,   // 默认显示条数
    MAX_LIMIT: 100,      // 最大显示条数
    ACTIVITY_LIMIT: 20,  // 活动记录显示条数
    KEYBOARD_LIMIT: 20,  // 键盘记录显示条数
    SUMMARY_LIMIT: 10,   // AI总结使用的数据条数
}
```

## 使用方法

### 1. 修改后端配置

编辑 `backend/config/config.yaml` 文件：

```bash
# 修改API密钥
vim backend/config/config.yaml

# 或使用环境变量指定配置文件
export CONFIG_PATH=/path/to/your/config.yaml
```

### 2. 修改前端配置

编辑 `frontend/web/config.js` 文件：

```javascript
// 修改API地址
const CONFIG = {
    API: {
        BASE_URL: 'http://your-server:8080/api/v1',
    },
    // ... 其他配置
};
```

### 3. Swift应用配置

在 `frontend/YAMLApp/Info.plist` 中添加：

```xml
<key>APIBaseURL</key>
<string>http://your-server:8080/api/v1</string>
```

## 配置验证

启动服务器时，系统会自动验证配置文件的有效性：

- 检查必需的配置项是否存在
- 验证数据类型和格式
- 创建必要的目录和文件

如果配置无效，服务器将显示详细的错误信息并退出。

## 环境变量支持

以下环境变量会覆盖配置文件中的设置：

- `CONFIG_PATH`: 配置文件路径
- `PORT`: 服务器端口
- `API_KEY`: AI API密钥
- `DB_PATH`: 数据库路径

## 安全注意事项

1. **API密钥安全**: 不要将包含真实API密钥的配置文件提交到版本控制系统
2. **文件权限**: 确保配置文件的权限设置合适（建议 600 或 644）
3. **敏感信息**: 考虑使用环境变量来存储敏感信息

## 示例配置

### 开发环境
```yaml
server:
  port: "8080"
  host: "localhost"
  
ai:
  gemini:
    api_key: "dev-api-key"
    base_url: "https://dev-api.example.com"
```

### 生产环境
```yaml
server:
  port: "80"
  host: "0.0.0.0"
  
ai:
  gemini:
    api_key: "${API_KEY}"  # 从环境变量读取
    base_url: "https://api.example.com"
    
logging:
  level: "warn"
  file_output: true
```

## 故障排除

### 常见问题

1. **配置文件未找到**
   - 检查文件路径是否正确
   - 确认文件存在且可读

2. **配置格式错误**
   - 验证YAML语法
   - 检查缩进和引号

3. **权限问题**
   - 确保应用有读取配置文件的权限
   - 检查数据目录的写入权限

### 调试模式

设置日志级别为 `debug` 来获取详细的配置加载信息：

```yaml
logging:
  level: "debug"
```