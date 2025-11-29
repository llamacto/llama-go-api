# Query Builder 使用文档

## 简介

Llamabase 提供了一个 Laravel 风格的查询构建器 (`pkg/database/dbx`)，它在 GORM 之上提供了更优雅、链式的 API。

## 特性

- **Laravel 风格**: 熟悉的链式调用方式
- **类型安全**: 基于 GORM，保持 Go 的类型安全
- **易用性**: 简化常见的数据库操作

---

## 快速开始

### 基础查询

```go
import "github.com/llamacto/llama-gin-kit/pkg/database/dbx"

// 查询所有用户
var users []User
err := dbx.Table("users").Get(&users)

// 查询单个用户
var user User
err := dbx.Table("users").Where("id = ?", 1).First(&user)

// 条件查询
var activeUsers []User
err := dbx.Table("users").
    Where("status = ?", "active").
    Order("created_at DESC").
    Limit(10).
    Get(&activeUsers)
```

### 使用模型

```go
// 基于模型查询
var user User
err := dbx.Model(&User{}).Where("email = ?", "user@example.com").First(&user)

// 等价于
err := dbx.Table("users").Where("email = ?", "user@example.com").First(&user)
```

---

## 查询方法

### WHERE 子句

```go
// 简单条件
dbx.Table("users").Where("age > ?", 18)

// 多个条件
dbx.Table("users").
    Where("age > ?", 18).
    Where("status = ?", "active")

// OR 条件
dbx.Table("users").
    Where("age > ?", 18).
    OrWhere("vip = ?", true)

// WHERE IN
dbx.Table("users").WhereIn("id", []int{1, 2, 3, 4, 5})

// WHERE NOT IN
dbx.Table("users").WhereNotIn("status", []string{"banned", "deleted"})

// Map 条件
dbx.Table("users").Where(map[string]interface{}{
    "status": "active",
    "role":   "admin",
})
```

### SELECT 字段

```go
// 选择特定字段
dbx.Table("users").
    Select("id", "name", "email").
    Get(&users)

// 聚合查询
var result struct {
    Count int
    Avg   float64
}
dbx.Table("users").
    Select("COUNT(*) as count", "AVG(age) as avg").
    First(&result)
```

### 排序

```go
// 升序
dbx.Table("users").Order("created_at ASC")

// 降序
dbx.Table("users").Order("created_at DESC")

// 多字段排序
dbx.Table("users").Order("status ASC, created_at DESC")
```

### 分页

```go
// 基础分页
dbx.Table("users").Offset(10).Limit(20).Get(&users)

// Laravel 风格分页
page := 1
perPage := 20
pagination, err := dbx.Table("users").Paginate(page, perPage, &users)
if err != nil {
    // 处理错误
}
// pagination 包含: Page, PerPage, Total
```

### JOIN 查询

```go
// INNER JOIN
dbx.Table("users").
    Join("JOIN orders ON users.id = orders.user_id").
    Where("orders.status = ?", "completed").
    Get(&results)

// LEFT JOIN
dbx.Table("users").
    LeftJoin("orders ON users.id = orders.user_id").
    Get(&results)

// RIGHT JOIN
dbx.Table("users").
    RightJoin("orders ON users.id = orders.user_id").
    Get(&results)
```

### GROUP BY & HAVING

```go
dbx.Table("orders").
    Select("user_id", "COUNT(*) as order_count", "SUM(amount) as total").
    Group("user_id").
    Having("COUNT(*) > ?", 5).
    Get(&results)
```

---

## 聚合方法

### Count

```go
total, err := dbx.Table("users").Where("status = ?", "active").Count()
```

### Exists

```go
exists, err := dbx.Table("users").Where("email = ?", "test@example.com").Exists()
```

### Value

```go
var email string
err := dbx.Table("users").Where("id = ?", 1).Value("email", &email)
```

---

## 写入操作

### 插入

```go
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}
err := dbx.Model(&User{}).Create(&user)
```

### 更新

```go
// 更新单个字段
err := dbx.Table("users").
    Where("id = ?", 1).
    Update("status", "active")

// 更新多个字段
updates := map[string]interface{}{
    "status":     "active",
    "updated_at": time.Now(),
}
err := dbx.Table("users").
    Where("id = ?", 1).
    Updates(updates)

// 使用结构体更新
user := User{Status: "active"}
err := dbx.Model(&User{}).
    Where("id = ?", 1).
    Updates(&user)
```

### 删除

```go
// 删除记录
err := dbx.Table("users").Delete(&User{}, 1)

// 条件删除
err := dbx.Model(&User{}).
    Where("status = ?", "banned").
    Delete(&User{})
```

---

## 高级用法

### 使用 Context

```go
ctx := context.Background()
err := dbx.Table("users").
    WithContext(ctx).
    Get(&users)
```

### 原始 DB 访问

```go
builder := dbx.Table("users")
gormDB := builder.DB() // 获取底层 GORM DB 实例
```

### 链式构建

```go
// 动态构建查询
builder := dbx.Table("users")

if keyword != "" {
    builder = builder.Where("name LIKE ?", "%"+keyword+"%")
}

if status != "" {
    builder = builder.Where("status = ?", status)
}

if sortBy != "" {
    builder = builder.Order(sortBy)
}

err := builder.Get(&users)
```

---

## 实战示例

### 示例 1: 用户列表带筛选

```go
func GetUserList(page, perPage int, keyword, status string) ([]*User, int64, error) {
    var users []*User
    
    builder := dbx.Table("users").
        Select("id", "name", "email", "status", "created_at")
    
    // 关键词搜索
    if keyword != "" {
        builder = builder.Where("name LIKE ? OR email LIKE ?", 
            "%"+keyword+"%", "%"+keyword+"%")
    }
    
    // 状态筛选
    if status != "" {
        builder = builder.Where("status = ?", status)
    }
    
    // 分页
    pagination, err := builder.
        Order("created_at DESC").
        Paginate(page, perPage, &users)
    
    if err != nil {
        return nil, 0, err
    }
    
    return users, pagination.Total, nil
}
```

### 示例 2: 统计查询

```go
func GetUserStats() (*UserStats, error) {
    var stats UserStats
    
    // 总用户数
    total, err := dbx.Table("users").Count()
    if err != nil {
        return nil, err
    }
    stats.Total = total
    
    // 活跃用户数
    active, err := dbx.Table("users").
        Where("status = ?", "active").
        Count()
    if err != nil {
        return nil, err
    }
    stats.Active = active
    
    // 今日新增
    today := time.Now().Format("2006-01-02")
    newToday, err := dbx.Table("users").
        Where("DATE(created_at) = ?", today).
        Count()
    if err != nil {
        return nil, err
    }
    stats.NewToday = newToday
    
    return &stats, nil
}
```

### 示例 3: 复杂关联查询

```go
func GetOrdersWithUser(status string) ([]OrderWithUser, error) {
    var results []OrderWithUser
    
    err := dbx.Table("orders").
        Select("orders.*, users.name as user_name, users.email as user_email").
        Join("JOIN users ON orders.user_id = users.id").
        Where("orders.status = ?", status).
        Order("orders.created_at DESC").
        Get(&results)
    
    return results, err
}
```

---

## 对比 GORM 原生用法

### GORM 原生

```go
var users []User
db.Where("status = ?", "active").
   Where("age > ?", 18).
   Order("created_at DESC").
   Limit(10).
   Find(&users)
```

### Llamabase Query Builder

```go
var users []User
dbx.Table("users").
    Where("status = ?", "active").
    Where("age > ?", 18).
    Order("created_at DESC").
    Limit(10).
    Get(&users)
```

**主要区别**:
- 使用 `Get()` 代替 `Find()`，语义更清晰
- 使用 `First()` 代替 `First()`，保持一致性
- 提供 Laravel 风格的 `Paginate()` 方法
- 简化的 `WhereIn()` 和 `WhereNotIn()` 方法

---

## 最佳实践

### 1. 始终检查错误

```go
var user User
err := dbx.Table("users").Where("id = ?", id).First(&user)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 处理记录不存在
    }
    // 处理其他错误
}
```

### 2. 使用 Context 控制超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := dbx.Table("users").WithContext(ctx).Get(&users)
```

### 3. 动态查询构建

```go
builder := dbx.Table("users")

// 根据条件动态添加 WHERE
if filters.Status != "" {
    builder = builder.Where("status = ?", filters.Status)
}

if filters.MinAge > 0 {
    builder = builder.Where("age >= ?", filters.MinAge)
}

err := builder.Get(&users)
```

### 4. 避免 N+1 查询

```go
// 不好的做法
var users []User
dbx.Table("users").Get(&users)
for _, user := range users {
    var orders []Order
    dbx.Table("orders").Where("user_id = ?", user.ID).Get(&orders)
    // ...
}

// 好的做法：使用 JOIN
var results []UserWithOrders
dbx.Table("users").
    LeftJoin("orders ON users.id = orders.user_id").
    Get(&results)
```

---

## API 参考

### 查询方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `Table(name)` | 指定表名 | `dbx.Table("users")` |
| `Model(value)` | 指定模型 | `dbx.Model(&User{})` |
| `Select(fields...)` | 选择字段 | `.Select("id", "name")` |
| `Where(query, args...)` | WHERE 条件 | `.Where("age > ?", 18)` |
| `OrWhere(query, args...)` | OR WHERE | `.OrWhere("vip = ?", true)` |
| `WhereIn(column, values)` | WHERE IN | `.WhereIn("id", []int{1,2,3})` |
| `WhereNotIn(column, values)` | WHERE NOT IN | `.WhereNotIn("status", []string{"banned"})` |
| `Order(value)` | 排序 | `.Order("created_at DESC")` |
| `Group(columns)` | GROUP BY | `.Group("user_id")` |
| `Having(query, args...)` | HAVING | `.Having("COUNT(*) > ?", 5)` |
| `Limit(limit)` | LIMIT | `.Limit(10)` |
| `Offset(offset)` | OFFSET | `.Offset(20)` |
| `Join(expr, args...)` | INNER JOIN | `.Join("JOIN orders...")` |
| `LeftJoin(expr, args...)` | LEFT JOIN | `.LeftJoin("orders...")` |
| `RightJoin(expr, args...)` | RIGHT JOIN | `.RightJoin("orders...")` |

### 执行方法

| 方法 | 说明 | 返回 |
|------|------|------|
| `Get(dest)` | 获取所有记录 | `error` |
| `First(dest)` | 获取第一条记录 | `error` |
| `Count()` | 计数 | `(int64, error)` |
| `Exists()` | 检查是否存在 | `(bool, error)` |
| `Value(column, dest)` | 获取单个字段值 | `error` |
| `Paginate(page, perPage, dest)` | 分页查询 | `(*Pagination, error)` |
| `Create(value)` | 创建记录 | `error` |
| `Update(column, value)` | 更新单个字段 | `error` |
| `Updates(values)` | 更新多个字段 | `error` |
| `Delete(value, conds...)` | 删除记录 | `error` |

### 工具方法

| 方法 | 说明 | 返回 |
|------|------|------|
| `WithContext(ctx)` | 设置 Context | `*Builder` |
| `DB()` | 获取底层 GORM DB | `*gorm.DB` |

---

## 常见问题

### Q: Query Builder 和 GORM 的关系？

A: Query Builder 是对 GORM 的封装，提供 Laravel 风格的 API。底层仍然使用 GORM，可以通过 `DB()` 方法访问原生 GORM 实例。

### Q: 如何执行原生 SQL？

A: 使用 GORM 原生方法：

```go
builder := dbx.Table("users")
db := builder.DB()
db.Raw("SELECT * FROM users WHERE custom_condition").Scan(&users)
```

### Q: 性能如何？

A: Query Builder 是轻量级封装，几乎没有性能损失。它使用 GORM 的 Session 机制，每次调用都是新的查询构建器实例。

---

## 未来计划

- [ ] 软删除支持
- [ ] 更多聚合方法 (Sum, Avg, Max, Min)
- [ ] 子查询支持
- [ ] 批量插入优化
- [ ] 事务包装器

---

**最后更新**: 2025-10-04
