# Llamabase - Laravel风格Go框架 TODO

## 项目概述
Llamabase 是一个受 Laravel 启发的 Go Web 框架，目标是提供优雅、直观的开发体验。

---

## ✅ 已完成 (Completed)

### 核心基础设施
- [x] 统一的配置加载系统 (`.env` 分层加载)
- [x] 配置缓存命令 (`config:cache`, `config:clear`)
- [x] 轻量级 IoC 容器 (Service Container)
- [x] Laravel 风格的查询构建器 (`pkg/database/dbx`)
- [x] CRUD 代码生成器 (类似 `php artisan make:model`)
- [x] 数据库迁移系统
- [x] 中间件系统 (JWT Auth, API Key Auth, Combined Auth)
- [x] 统一响应格式
- [x] 用户认证系统 (User, Organization, Team, API Key)

---

## 🚧 进行中 (In Progress)

### 文档与测试
- [ ] **编写 Query Builder 使用文档** (`pkg/database/dbx`)
  - 说明 Laravel 风格的链式调用
  - 提供常见查询示例
  - 对比 GORM 原生用法

- [ ] **创建单元测试**
  - Container 测试
  - Query Builder 测试
  - 配置系统测试
  - 中间件测试

- [ ] **创建集成测试**
  - API 端点测试
  - 认证流程测试
  - 配置分层测试

---

## 📋 近期规划 (Near Term)

### 1. Artisan 风格命令行工具 (Priority: High)
**目标**: 统一的 CLI 入口，类似 Laravel Artisan

```bash
llamabase serve                    # 启动服务器
llamabase migrate                  # 运行迁移
llamabase migrate:rollback         # 回滚迁移
llamabase make:model User          # 生成模型
llamabase make:controller User     # 生成控制器
llamabase config:cache             # 缓存配置
llamabase config:clear             # 清除配置缓存
llamabase route:list               # 列出所有路由
llamabase tinker                   # 交互式 REPL (可选)
```

**实现步骤**:
- [ ] 创建 `cmd/llamabase/main.go` 作为统一入口
- [ ] 整合现有的 generator, migrate, tools 命令
- [ ] 添加命令行参数解析和帮助文档
- [ ] 支持自定义命令注册

### 2. Service Provider 系统 (Priority: High)
**目标**: Laravel 风格的服务提供者，实现模块化启动

```go
// 示例：DatabaseServiceProvider
type DatabaseServiceProvider struct{}

func (p *DatabaseServiceProvider) Register(app *container.Container) {
    app.Bind(container.ServiceDB, func() (any, error) {
        return database.InitDB(config.GlobalConfig.Database)
    })
}

func (p *DatabaseServiceProvider) Boot() {
    // 数据库启动后的初始化逻辑
}
```

**实现步骤**:
- [ ] 定义 `ServiceProvider` 接口
- [ ] 实现核心 Provider (Database, Email, Logger, JWT)
- [ ] 在 `cmd/server/main.go` 中注册和启动 Providers
- [ ] 支持延迟加载（deferred providers）

### 3. 路由系统优化 (Priority: Medium)
**目标**: Laravel 风格的路由定义

```go
// 目标 API
Route.Get("/users", userHandler.List).Middleware(auth.JWT())
Route.Post("/users", userHandler.Create)
Route.Group("/api/v1", func(r *RouteGroup) {
    r.Middleware(cors.Default())
    r.Get("/posts", postHandler.List)
})
```

**实现步骤**:
- [ ] 创建 `pkg/routing` 包
- [ ] 实现链式路由注册
- [ ] 支持路由组和中间件
- [ ] 生成路由列表命令 (`route:list`)

### 4. Facade 系统 (Priority: Medium)
**目标**: 类似 Laravel Facade 的全局访问器

```go
// 使用示例
DB.Table("users").Where("id", 1).First(&user)
Cache.Set("key", "value", 10*time.Minute)
Mail.To("user@example.com").Send(new(WelcomeMail))
```

**实现步骤**:
- [ ] 设计 Facade 接口
- [ ] 实现核心 Facades (DB, Cache, Log, Mail)
- [ ] 从容器中解析实例

---

## 🔮 长期规划 (Future Ideas)

### 1. 队列系统 (Queue & Jobs)
- [ ] 定义 Job 接口
- [ ] 实现队列驱动 (Redis, Database, Memory)
- [ ] 队列工作进程 (`llamabase queue:work`)
- [ ] 失败任务重试机制

### 2. 事件系统 (Events & Listeners)
- [ ] 事件分发器
- [ ] 事件监听器注册
- [ ] 异步事件处理

### 3. 验证系统 (Validation)
- [ ] Laravel 风格的验证规则
- [ ] 自定义验证器
- [ ] 表单请求验证

```go
// 目标 API
type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min:8"`
}
```

### 4. ORM 增强
- [ ] 模型关系定义 (HasOne, HasMany, BelongsTo, ManyToMany)
- [ ] Eager Loading
- [ ] 软删除
- [ ] 模型事件 (Creating, Created, Updating, Updated)

### 5. 缓存系统
- [ ] 统一缓存接口
- [ ] 多驱动支持 (Redis, Memory, File)
- [ ] 缓存标签 (Cache Tags)

### 6. 任务调度 (Task Scheduling)
- [ ] Cron 风格的任务调度
- [ ] `llamabase schedule:run`

### 7. 多租户支持
- [ ] 租户配置覆盖
- [ ] 租户数据隔离
- [ ] 动态数据库切换

### 8. API 资源 (API Resources)
- [ ] 响应转换器
- [ ] 资源集合
- [ ] 条件字段

---

## 📚 文档计划

### 使用文档
- [ ] 快速入门指南
- [ ] 配置系统详解
- [ ] 路由与中间件
- [ ] 数据库与 ORM
- [ ] 认证与授权
- [ ] 命令行工具
- [ ] 最佳实践

### API 文档
- [ ] 完善 Swagger 注释
- [ ] 生成 API 文档站点

---

## 🧪 测试与质量保证

### 单元测试
- [ ] Container 包
- [ ] Config 包
- [ ] Database/DBX 包
- [ ] Middleware 包
- [ ] JWT 包

### 集成测试
- [ ] 用户注册登录流程
- [ ] API Key 认证
- [ ] 组织和团队管理
- [ ] 配置缓存机制

### 性能测试
- [ ] 基准测试 (Benchmark)
- [ ] 压力测试

---

## 🎯 优化目标

### 代码质量
- [ ] 添加 golangci-lint 配置
- [ ] 代码覆盖率达到 80%+
- [ ] 统一代码风格

### 性能优化
- [ ] 数据库连接池优化
- [ ] 响应缓存
- [ ] 减少内存分配

### 开发体验
- [ ] 更详细的错误信息
- [ ] 开发模式热重载 (已有 Air)
- [ ] 交互式调试工具

---

## 🚀 发布计划

### v0.1.0 (当前)
- 基础框架搭建
- 核心功能实现

### v0.2.0 (下一版本)
- Artisan CLI 完成
- Service Provider 系统
- 完整测试覆盖
- 文档完善

### v1.0.0 (稳定版)
- 所有核心功能稳定
- 生产环境验证
- 性能优化完成
- 完整文档

---

**最后更新**: 2025-10-04  
**维护者**: Llamacto Team
