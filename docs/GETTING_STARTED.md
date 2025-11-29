# Llamabase 快速入门

欢迎使用 Llamabase - 一个优雅的 Go Web 框架！

## 🎯 什么是 Llamabase？

Llamabase 是一个受 Laravel 启发的 Go Web 框架，旨在为 Go 开发者提供类似 Laravel 的优雅开发体验。

### 核心特性

- **Laravel 风格 API**: 熟悉的链式调用和语义化方法
- **强大的配置系统**: 分层环境变量和配置缓存
- **IoC 容器**: 优雅的依赖注入
- **Query Builder**: Laravel 风格的数据库查询
- **代码生成器**: 快速生成样板代码
- **完整的认证系统**: JWT 和 API Key 双重支持

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6.0+ (可选)

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/llamacto/llamabase.git
cd llamabase
```

#### 2. 安装依赖

```bash
go mod download
```

#### 3. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置数据库连接：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
DB_NAME=llamabase

# JWT 配置
JWT_SECRET=your_secret_key_here
```

#### 4. 创建数据库

```bash
# 登录 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE llamabase;
```

#### 5. 运行迁移

```bash
make migrate
```

#### 6. 启动服务

```bash
# 开发模式（热重载）
make air

# 或者普通启动
make run
```

#### 7. 访问 API

打开浏览器访问：
- **API 首页**: http://localhost:6066/
- **Swagger 文档**: http://localhost:6066/swagger/index.html
- **健康检查**: http://localhost:6066/v1/health/status

---

## 📚 基础教程

### 1. 创建一个新模块

使用代码生成器快速创建 CRUD 模块：

```bash
make generate model=Product table=products package=product
```

这会生成：
- `app/product/model.go` - 数据模型
- `app/product/service.go` - 业务逻辑
- `app/product/handler.go` - HTTP 处理器
- `app/product/repository.go` - 数据访问层

### 2. 使用 Query Builder

```go
import "github.com/llamacto/llama-gin-kit/pkg/database/dbx"

// 查询所有活跃用户
var users []User
err := dbx.Table("users").
    Where("status = ?", "active").
    Order("created_at DESC").
    Limit(10).
    Get(&users)

// 分页查询
pagination, err := dbx.Table("users").
    Paginate(page, perPage, &users)
```

### 3. 使用 IoC 容器

```go
import "github.com/llamacto/llama-gin-kit/pkg/container"

// 注册服务
container.App().Set("myService", myServiceInstance)

// 解析服务
service, err := container.App().Resolve("myService")

// 类型安全解析
service := container.MustResolveAs[*MyService]("myService")
```

### 4. 创建 API 接口

#### 定义模型

```go
// app/product/model.go
type Product struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Name        string    `json:"name" binding:"required"`
    Description string    `json:"description"`
    Price       float64   `json:"price" binding:"required,gt=0"`
    Stock       int       `json:"stock" binding:"required,gte=0"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### 实现 Repository

```go
// app/product/repository.go
func (r *ProductRepository) Create(ctx context.Context, product *Product) error {
    return dbx.Model(&Product{}).
        WithContext(ctx).
        Create(product)
}

func (r *ProductRepository) List(ctx context.Context, page, perPage int) ([]*Product, int64, error) {
    var products []*Product
    pagination, err := dbx.Table("products").
        WithContext(ctx).
        Order("created_at DESC").
        Paginate(page, perPage, &products)
    
    if err != nil {
        return nil, 0, err
    }
    
    return products, pagination.Total, nil
}
```

#### 实现 Handler

```go
// app/product/handler.go
func (h *ProductHandler) Create(c *gin.Context) {
    var product Product
    if err := c.ShouldBindJSON(&product); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid input")
        return
    }
    
    if err := h.service.Create(c.Request.Context(), &product); err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to create product")
        return
    }
    
    response.Success(c, product)
}
```

#### 注册路由

```go
// routes/v1/routes.go
productRepo := product.NewProductRepository()
productService := product.NewProductService(productRepo)
productHandler := product.NewProductHandler(productService)

v1.POST("/products", productHandler.Create)
v1.GET("/products", productHandler.List)
v1.GET("/products/:id", productHandler.Get)
v1.PUT("/products/:id", productHandler.Update)
v1.DELETE("/products/:id", productHandler.Delete)
```

### 5. 使用中间件

```go
// 需要 JWT 认证的路由
protectedGroup := v1.Group("/admin")
protectedGroup.Use(pkgmiddleware.JWTAuth())
{
    protectedGroup.GET("/dashboard", dashboardHandler)
}

// 支持 JWT 或 API Key 认证
combinedAuth := middleware.CombinedAuth(apiKeyService)
v1.GET("/data", combinedAuth, dataHandler)
```

### 6. 配置管理

```go
import "github.com/llamacto/llama-gin-kit/config"

// 加载配置
cfg, err := config.Load()

// 访问配置
dbHost := cfg.Database.Host
jwtSecret := cfg.JWT.Secret

// 缓存配置（提升性能）
go run cmd/tools/main.go -tool=config-cache

// 清除缓存
go run cmd/tools/main.go -tool=config-clear
```

---

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./pkg/container/
go test -v ./pkg/response/

# 运行基准测试
go test -bench=. ./pkg/container/
```

### 编写测试

```go
func TestProductService_Create(t *testing.T) {
    // 准备测试数据
    product := &Product{
        Name:  "Test Product",
        Price: 99.99,
        Stock: 10,
    }
    
    // 执行测试
    err := service.Create(context.Background(), product)
    
    // 验证结果
    if err != nil {
        t.Fatalf("Failed to create product: %v", err)
    }
    
    if product.ID == 0 {
        t.Error("Product ID should be set")
    }
}
```

---

## 📖 API 示例

### 用户注册

```bash
curl -X POST http://localhost:6066/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### 用户登录

```bash
curl -X POST http://localhost:6066/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 访问受保护的资源

```bash
curl -X GET http://localhost:6066/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 使用 API Key

```bash
curl -X GET http://localhost:6066/v1/protected \
  -H "X-API-Key: YOUR_API_KEY"
```

---

## 🛠️ 常用命令

### Make 命令

```bash
make build       # 编译项目
make run         # 运行服务
make air         # 热重载模式
make test        # 运行测试
make migrate     # 运行数据库迁移
make swagger     # 生成 Swagger 文档
make fmt         # 格式化代码
make clean       # 清理构建文件
```

### 代码生成

```bash
# 生成完整模块
make generate model=User table=users

# 自定义包名
make generate model=Product package=prod

# 强制覆盖
make generate model=Order force=true
```

---

## 📂 项目结构

```
llamabase/
├── cmd/                    # 命令行入口
│   ├── server/            # HTTP 服务器
│   ├── migrate/           # 数据库迁移
│   ├── generator/         # 代码生成器
│   └── tools/             # 工具命令
├── app/                    # 业务模块
│   ├── user/              # 用户模块
│   ├── organization/      # 组织模块
│   ├── team/              # 团队模块
│   └── apikey/            # API Key 模块
├── config/                 # 配置管理
├── middleware/             # 中间件
├── pkg/                    # 工具包
│   ├── container/         # IoC 容器
│   ├── database/          # 数据库
│   │   └── dbx/          # Query Builder
│   ├── response/          # 响应助手
│   ├── jwt/               # JWT 认证
│   └── email/             # 邮件服务
├── routes/                 # 路由定义
├── docs/                   # 文档
└── storage/                # 存储目录
    └── framework/
        └── cache/         # 配置缓存
```

---

## 🎓 学习资源

### 文档

- **Query Builder 文档**: `docs/query-builder.md`
- **TODO 规划**: `docs/TODO.md`
- **项目评审**: `docs/PROJECT_REVIEW.md`
- **API 示例**: `docs/api_examples.md`

### 代码示例

查看 `app/` 目录下的各个模块，了解完整的实现示例：
- `app/user/` - 用户管理
- `app/organization/` - 组织管理
- `app/team/` - 团队管理

---

## 💡 最佳实践

### 1. 使用 Repository 模式

将数据访问逻辑封装在 Repository 中：

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
}
```

### 2. 使用 Service 层

业务逻辑放在 Service 层：

```go
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // 验证邮箱是否已存在
    exists, _ := s.repo.FindByEmail(ctx, req.Email)
    if exists != nil {
        return nil, errors.New("email already exists")
    }
    
    // 创建用户
    user := &User{
        Email:    req.Email,
        Password: hashPassword(req.Password),
        Name:     req.Name,
    }
    
    return user, s.repo.Create(ctx, user)
}
```

### 3. 统一错误处理

使用统一的响应格式：

```go
response.Success(c, data)           // 成功响应
response.Error(c, 400, "message")   // 错误响应
```

### 4. 使用 Context

传递请求上下文：

```go
ctx := c.Request.Context()
user, err := userService.Get(ctx, userID)
```

---

## 🐛 故障排除

### 数据库连接失败

检查 `.env` 配置：
```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
DB_NAME=llamabase
```

确保数据库已创建并可访问。

### 端口被占用

修改 `.env` 中的端口：
```env
SERVER_PORT=8080
```

### JWT 认证失败

确保设置了 JWT Secret：
```env
JWT_SECRET=your_secret_key_here
```

---

## 🤝 获取帮助

- **文档**: 查看 `docs/` 目录
- **示例**: 参考 `app/` 模块代码
- **Issues**: 提交 GitHub Issue

---

## 🎉 下一步

1. 阅读 **Query Builder 文档**，掌握数据库查询
2. 查看 **TODO 规划**，了解路线图
3. 阅读 **项目评审报告**，深入理解架构
4. 开始构建你的第一个模块！

---

**Happy Coding with Llamabase! 🦙**
