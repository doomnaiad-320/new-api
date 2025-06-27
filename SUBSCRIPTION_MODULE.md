# 🎉 订阅计费模块

## 📋 概述

订阅计费模块是对现有 New API 系统的重要扩展，实现了**多重计费模式**的无缝集成。用户可以购买订阅套餐享受优惠价格，当订阅配额用完时自动切换到原有的 token/次数计费模式，确保服务不中断。

## ✨ 核心特性

### 🔄 多重计费模式
- **订阅优先**: 优先使用订阅套餐中的配额
- **自动切换**: 订阅配额耗尽时自动切换到按量计费
- **无缝体验**: 用户无感知的计费模式切换

### 📦 灵活套餐配置
- **多模型支持**: 每个套餐可包含多个 AI 模型的不同配额
- **自定义价格**: 灵活设置套餐价格和有效期
- **动态管理**: 支持套餐的启用/禁用、编辑删除

### 📊 实时监控统计
- **配额跟踪**: 实时监控各模型配额使用情况
- **预警提醒**: 配额不足时自动发送预警通知
- **详细报表**: 收入统计、使用分析、趋势报告

### 🎨 完整管理界面
- **管理后台**: 套餐管理、用户订阅、统计报表
- **用户界面**: 套餐购买、配额查看、使用记录
- **响应式设计**: 支持多种设备和屏幕尺寸

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 8.0+ 或 PostgreSQL 12+
- Redis 6.0+ (可选)

### 安装部署

#### 1. 源码部署
```bash
# 克隆代码
git clone https://github.com/your-repo/new-api.git
cd new-api

# 构建前端
cd web
npm install
npm run build
cd ..

# 构建后端
go build -o new-api

# 运行
./new-api
```

#### 2. Docker 部署
```bash
# 使用最新的订阅模块版本
docker run -d \
  --name new-api \
  -p 3000:3000 \
  -e SQL_DSN="your_database_connection" \
  doomnaiad/new-api:subscription-latest
```

#### 3. Docker Compose
```yaml
version: '3.8'
services:
  new-api:
    image: doomnaiad/new-api:subscription-latest
    ports:
      - "3000:3000"
    environment:
      - SQL_DSN=mysql://user:password@mysql:3306/oneapi
      - REDIS_CONN_STRING=redis://redis:6379
    depends_on:
      - mysql
      - redis
  
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=your_password
      - MYSQL_DATABASE=oneapi
    volumes:
      - mysql_data:/var/lib/mysql
  
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

### 数据库迁移

系统启动时会自动执行数据库迁移，创建以下新表：
- `subscription_plans`: 订阅套餐表
- `user_subscriptions`: 用户订阅表  
- `subscription_usages`: 订阅使用记录表

## 📖 使用指南

### 管理员操作

#### 1. 创建订阅套餐
```bash
# 访问管理后台
http://your-domain/console/subscription

# 创建套餐示例
{
  "name": "基础套餐",
  "description": "适合轻度使用的用户",
  "price": 10.00,
  "duration": 30,
  "model_quotas": {
    "gpt-4": 100,
    "claude-3.5-sonnet": 50,
    "gemini-1.5-pro": 200
  }
}
```

#### 2. 查看用户订阅
```bash
# 用户订阅管理
http://your-domain/console/subscription/users

# 统计报表
http://your-domain/console/subscription/stats
```

### 用户操作

#### 1. 购买订阅套餐
```bash
# 访问购买页面
http://your-domain/subscription/purchase

# 选择套餐并完成支付
```

#### 2. 查看配额使用
```bash
# API 查询当前配额
GET /api/subscription/quotas

# 响应示例
{
  "success": true,
  "data": {
    "quotas": {
      "gpt-4": {
        "total": 100,
        "used": 25,
        "remaining": 75
      }
    }
  }
}
```

## 🔧 API 接口

### 套餐管理
```bash
# 获取所有套餐
GET /api/subscription/plans

# 创建套餐 (管理员)
POST /api/subscription/admin/plans

# 更新套餐 (管理员)
PUT /api/subscription/admin/plans/:id

# 删除套餐 (管理员)
DELETE /api/subscription/admin/plans/:id
```

### 用户订阅
```bash
# 购买订阅
POST /api/subscription/purchase

# 查看我的订阅
GET /api/subscription/my

# 查看激活订阅
GET /api/subscription/active

# 查看配额状态
GET /api/subscription/quotas
```

### 统计报表
```bash
# 使用统计
GET /api/subscription/stats

# 系统统计 (管理员)
GET /api/subscription/admin/system-stats

# 生成报表 (管理员)
GET /api/subscription/admin/report
```

## 🎯 配置说明

### 环境变量
```bash
# 数据库连接
SQL_DSN="mysql://user:password@localhost:3306/oneapi"

# Redis 连接 (可选)
REDIS_CONN_STRING="redis://localhost:6379"

# 会话密钥
SESSION_SECRET="your_session_secret"

# 订阅监控间隔 (小时)
SUBSCRIPTION_MONITOR_INTERVAL=1

# 使用记录保留天数
USAGE_RETENTION_DAYS=90
```

### 套餐配置示例
```json
{
  "plans": [
    {
      "name": "入门套餐",
      "price": 9.9,
      "duration": 30,
      "model_quotas": {
        "gpt-3.5-turbo": 1000,
        "gpt-4o-mini": 500
      }
    },
    {
      "name": "专业套餐", 
      "price": 29.9,
      "duration": 30,
      "model_quotas": {
        "gpt-4": 200,
        "claude-3.5-sonnet": 100,
        "gemini-1.5-pro": 300
      }
    }
  ]
}
```

## 🔍 监控和维护

### 自动监控
- **配额监控**: 每小时检查用户配额使用情况
- **过期清理**: 自动更新过期订阅状态
- **使用记录**: 定期清理旧的使用记录

### 手动维护
```bash
# 手动触发监控
POST /api/subscription/admin/monitor

# 清理过期订阅
POST /api/subscription/admin/cleanup

# 查看系统状态
GET /api/subscription/admin/system-stats
```

## 🐛 故障排除

### 常见问题

1. **订阅配额不生效**
   - 检查订阅是否在有效期内
   - 确认套餐状态为启用
   - 查看模型名称是否匹配

2. **计费切换异常**
   - 检查用户余额是否充足
   - 确认原有计费系统正常
   - 查看错误日志

3. **前端页面异常**
   - 清除浏览器缓存
   - 检查 API 接口连通性
   - 确认用户权限

### 日志查看
```bash
# 查看订阅相关日志
grep "subscription" /path/to/logs/app.log

# 查看配额消费日志
grep "quota" /path/to/logs/app.log
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发环境
```bash
# 安装依赖
go mod download
cd web && npm install

# 运行开发服务器
go run main.go

# 前端开发
cd web && npm run dev
```

### 测试
```bash
# 运行测试
go test ./...

# 前端测试
cd web && npm test
```

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢所有贡献者和用户的支持！

---

**🎉 享受更智能的 AI API 计费体验！**
