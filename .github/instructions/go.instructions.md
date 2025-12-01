---
applyTo: '**/*.go'
---

## 1. 测试驱动开发（TDD）

### 强制要求
- **测试先行**: 在编写任何实现代码之前，必须先编写测试
- **测试文件**: 每个 `.go` 文件必须有对应的 `_test.go` 文件
- **测试覆盖率**: 最低 80%，核心业务逻辑要求 100%
- **表格驱动测试**: 使用表格驱动测试模式处理多个测试用例

### 测试命名规范
- 测试函数以 `Test` 开头，后跟被测试的函数名
- 基准测试以 `Benchmark` 开头
- 示例测试以 `Example` 开头
- 子测试使用 `t.Run()` 并提供清晰的描述

### 必需的测试类型
```go
// 单元测试
func TestFunctionName(t *testing.T)

// 表格驱动测试
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        // 测试用例
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}

// 基准测试
func BenchmarkFunctionName(b *testing.B)

// 示例测试（文档）
func ExampleFunctionName()
```

## 2. 命名规范

### 包名（Package）
- **简短小写**: 使用简短、小写、单数形式
- **无下划线**: 不使用下划线或驼峰命名
- **有意义**: 包名应该清晰表达其用途
- **避免通用名**: 不使用 `util`、`common`、`base` 等过于通用的名称

**正确示例**: `http`, `json`, `user`, `auth`, `storage`  
**错误示例**: `HTTP`, `user_service`, `utils`, `myPackage`

### 变量和函数命名
- **导出标识符**: 大写字母开头（public）
- **未导出标识符**: 小写字母开头（private）
- **驼峰命名**: 使用 `MixedCaps` 或 `mixedCaps`
- **缩略词**: 保持大写或全部小写（如 `HTTP`, `ID`, `URL`）

**正确示例**:
```go
var userID int           // 正确：ID 全大写
var httpClient *Client   // 正确：HTTP 全大写
var serverURL string     // 正确：URL 全大写

func GetUserByID(id int) // 导出函数
func parseHTTPRequest()  // 未导出函数
```

**错误示例**:
```go
var userId int          // 错误：Id 应该全大写
var HTTPClient *Client  // 错误：导出的应该是 HTTPClient 或 Client
var server_url string   // 错误：不使用下划线
```

### 接口命名
- **单方法接口**: 以 `-er` 结尾
- **多方法接口**: 使用名词或名词短语
- **避免 `I` 前缀**: Go 不使用 `IReader` 这样的命名

**正确示例**: `Reader`, `Writer`, `Formatter`, `UserRepository`  
**错误示例**: `IReader`, `ReaderInterface`, `ReadWriter` (除非确实有 Read 和 Write 方法)

### 获取器和设置器
- **获取器**: 直接使用名词，不加 `Get` 前缀
- **设置器**: 使用 `Set` 前缀

**正确示例**:
```go
func (u *User) Name() string      // 获取器
func (u *User) SetName(name string) // 设置器
```

**错误示例**:
```go
func (u *User) GetName() string  // 错误：不需要 Get 前缀
```

### 常量命名
- 使用驼峰命名，不使用全大写蛇形命名
- 导出的常量大写开头，未导出的小写开头

**正确示例**:
```go
const MaxConnections = 100
const defaultTimeout = 30 * time.Second
```

**错误示例**:
```go
const MAX_CONNECTIONS = 100  // 错误：不使用蛇形命名
```

## 3. 代码组织

### 文件结构顺序
1. Package 声明
2. Import 语句（分组：标准库、第三方库、本项目）
3. 常量声明
4. 变量声明
5. 类型定义
6. init 函数（如果需要）
7. 主要函数和方法
8. 辅助/私有函数

### Import 语句规范
- **分组**: 标准库、第三方库、本地包，组之间空行分隔
- **禁止点导入**: 不使用 `.` 导入
- **禁止空白导入**: 除非必要（如数据库驱动），需添加注释说明
- **使用 goimports**: 自动格式化和管理 import

**正确示例**:
```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"

    "github.com/yourorg/yourproject/internal/auth"
    "github.com/yourorg/yourproject/pkg/errors"
)
```

### 包组织原则
- **按功能组织**: 不是按类型（不要创建 `models/`、`controllers/` 包）
- **小而专注**: 每个包应该有明确、单一的职责
- **internal 包**: 使用 `internal/` 限制包的可见性
- **避免循环依赖**: 设计清晰的依赖关系图

## 4. 接口设计

### 接口定义原则
- **小接口**: 优先设计单方法接口
- **按需定义**: 在使用处定义接口，而非提供处
- **组合接口**: 通过嵌入组合小接口成大接口
- **避免空接口**: 谨慎使用 `interface{}` 或 `any`

**正确示例**:
```go
// 小而专注的接口
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 通过组合构建
type ReadWriter interface {
    Reader
    Writer
}
```

### 接受接口，返回结构体
- **函数参数**: 接受接口类型，提供灵活性
- **返回值**: 返回具体类型，便于使用和文档

**正确示例**:
```go
// 接受接口
func ProcessData(r io.Reader) (*Result, error) {
    // 实现
}

// 返回具体类型
func NewUser(name string) *User {
    return &User{Name: name}
}
```

## 5. 错误处理

### 错误处理原则
- **显式检查**: 每个错误都必须检查，不能忽略
- **及时处理**: 在错误发生处或附近处理
- **错误包装**: 使用 `fmt.Errorf("%w", err)` 包装错误，添加上下文
- **错误类型**: 定义有意义的错误类型用于判断

### 错误处理规范
```go
// ✅ 正确：显式检查每个错误
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ✅ 正确：自定义错误类型
type ValidationError struct {
    Field string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ✅ 正确：使用 errors.Is 和 errors.As
if errors.Is(err, ErrNotFound) {
    // 处理未找到的情况
}

var validErr *ValidationError
if errors.As(err, &validErr) {
    // 处理验证错误
}
```

### 错误命名
- 错误变量以 `Err` 开头
- 错误类型以 `Error` 结尾

**示例**: `ErrNotFound`, `ErrInvalidInput`, `ValidationError`

### Panic 和 Recover
- **避免 panic**: 仅在不可恢复的错误时使用
- **库代码禁用**: 库代码不应 panic，应返回 error
- **main 和 init**: 仅在 main 和 init 函数中可以 panic
- **recover**: 仅在必要时使用，通常在 defer 中

## 6. 并发编程

### Goroutine 管理
- **避免泄漏**: 确保每个 goroutine 都有退出路径
- **使用 context**: 传递取消信号和超时控制
- **使用 WaitGroup**: 等待一组 goroutine 完成
- **限制数量**: 避免创建无限制的 goroutine

### Channel 使用规范
- **所有权**: 只有发送方应该关闭 channel
- **缓冲区大小**: 明确设置缓冲区大小，避免默认值
- **select 语句**: 使用 default 避免阻塞（如果适用）
- **单向 channel**: 使用单向 channel 限制操作

**正确示例**:
```go
// 正确的 goroutine 管理
func ProcessItems(ctx context.Context, items []Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := processItem(ctx, item); err != nil {
                errChan <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errChan)
    
    // 检查错误
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 并发安全
- **共享内存**: 使用 `sync.Mutex` 或 `sync.RWMutex` 保护
- **通信优先**: "通过通信共享内存，而非通过共享内存通信"
- **race 检测**: 始终运行 `go test -race`
- **sync.Once**: 用于一次性初始化

### Context 使用
- **第一个参数**: context 应该是函数的第一个参数
- **命名为 ctx**: 统一命名为 `ctx`
- **不存储**: 不在结构体中存储 context
- **传递**: 显式传递，不使用全局 context

**正确示例**:
```go
func FetchUser(ctx context.Context, userID int) (*User, error) {
    // 实现
}
```

## 7. 函数和方法设计

### 函数签名
- **参数顺序**: context, 必需参数, 可选参数
- **返回值**: (结果, error) 或 (结果1, 结果2, error)
- **命名返回值**: 仅在有助于文档说明时使用
- **参数数量**: 超过 3 个参数考虑使用配置结构体

### 方法接收者
- **指针接收者**: 需要修改接收者、接收者较大、或一致性要求
- **值接收者**: 小型不可变结构体、基本类型
- **命名**: 使用类型首字母或简短缩写，保持一致

**正确示例**:
```go
type User struct {
    Name string
    Age  int
}

func (u *User) SetName(name string) {  // 指针接收者：修改
    u.Name = name
}

func (u User) IsAdult() bool {  // 值接收者：只读
    return u.Age >= 18
}
```

### 函数选项模式
对于复杂配置，使用函数选项模式：

```go
type ServerOptions struct {
    Port    int
    Timeout time.Duration
}

type ServerOption func(*ServerOptions)

func WithPort(port int) ServerOption {
    return func(o *ServerOptions) {
        o.Port = port
    }
}

func NewServer(opts ...ServerOption) *Server {
    options := &ServerOptions{
        Port:    8080,  // 默认值
        Timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(options)
    }
    return &Server{options: options}
}
```

## 8. 类型和结构体

### 结构体定义
- **零值可用**: 设计结构体使其零值是有用的
- **字段顺序**: 按重要性或逻辑分组，考虑内存对齐
- **嵌入**: 谨慎使用嵌入，优先使用组合
- **标签**: 使用结构体标签时保持一致的格式

**正确示例**:
```go
type Config struct {
    // 基本配置
    Host string
    Port int
    
    // 超时配置
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    
    // 可选配置
    Logger *log.Logger  // 可以为 nil，使用默认值
}
```

### 类型定义
- **有意义的类型**: 为特定用途创建类型别名
- **避免过度使用**: 不为所有基础类型创建别名
- **文档说明**: 为自定义类型添加文档

```go
// UserID 表示用户的唯一标识符
type UserID int64

// Email 表示验证过的电子邮件地址
type Email string
```

## 9. 文档和注释

### 包文档
- 每个包必须有包级文档注释
- 包文档在 `package` 声明之前
- 使用完整句子，以包名开头

**正确示例**:
```go
// Package auth 提供用户认证和授权功能。
//
// 本包实现了 JWT 令牌生成、验证和刷新机制。
// 支持多种认证策略，包括基本认证和 OAuth2。
package auth
```

### 导出标识符文档
- **强制要求**: 所有导出的类型、函数、常量、变量必须有文档
- **完整句子**: 使用完整句子，以标识符名称开头
- **说明用途**: 解释是什么、为什么，而非怎么做

**正确示例**:
```go
// User 表示系统中的一个用户账户。
type User struct {
    Name string
    Email string
}

// NewUser 创建并返回一个新的用户实例。
// 如果提供的邮箱格式无效，返回 ErrInvalidEmail。
func NewUser(name, email string) (*User, error) {
    // 实现
}
```

### 注释风格
- **为什么而非什么**: 解释设计决策和非显而易见的逻辑
- **TODO 注释**: 使用 `TODO(username): description` 格式
- **避免无用注释**: 不重复代码已经表达的内容

**正确示例**:
```go
// 使用二次指数退避避免过载服务器
time.Sleep(backoff)

// TODO(alice): 添加重试逻辑处理临时网络错误
```

**错误示例**:
```go
// 设置 i 为 0
i := 0  // 无意义的注释
```

## 10. 代码质量要求

### 格式化
- **强制使用**: `gofmt -s` 或 `goimports`
- **提交前**: 所有代码必须格式化
- **编辑器集成**: 配置编辑器保存时自动格式化

### 静态检查
- **go vet**: 必须通过 `go vet ./...`
- **golangci-lint**: 必须通过 linter 检查
- **staticcheck**: 推荐使用额外的静态分析工具

### 性能
- **基准测试**: 性能关键代码必须有基准测试
- **避免过早优化**: 先保证正确性和可读性
- **pprof**: 使用 pprof 分析性能瓶颈
- **内存分配**: 注意减少不必要的堆分配

### 安全性
- **输入验证**: 验证所有外部输入
- **SQL 注入**: 使用参数化查询
- **敏感信息**: 不在代码中硬编码密钥、密码
- **依赖扫描**: 定期扫描依赖安全漏洞

## 11. 最佳实践

### 使用内置函数和标准库
- 优先使用标准库而非第三方库
- 熟悉并使用 `strings`、`bytes`、`io` 等标准包
- 了解内置函数如 `copy`、`append`、`make`、`len`

### 避免的反模式
```go
// ❌ 避免：忽略错误
result, _ := DoSomething()

// ❌ 避免：过度使用空接口
func Process(data interface{}) // 类型不安全

// ❌ 避免：全局变量
var GlobalConfig Config

// ❌ 避免：init 函数中的复杂逻辑
func init() {
    // 避免在 init 中进行网络调用、文件操作等
}

// ❌ 避免：过长的函数
func VeryLongFunction() {
    // 超过 50 行的函数应该考虑拆分
}
```

### 推荐的模式
```go
// ✅ 显式错误处理
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// ✅ 使用具体类型
func Process(data *SpecificType) error

// ✅ 依赖注入
type Service struct {
    repo Repository
    logger Logger
}

func NewService(repo Repository, logger Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}

// ✅ 早返回（guard clauses）
func ValidateUser(user *User) error {
    if user == nil {
        return ErrNilUser
    }
    if user.Name == "" {
        return ErrEmptyName
    }
    if user.Age < 0 {
        return ErrInvalidAge
    }
    return nil
}
```

## 12. 代码审查检查清单

在提交代码前，确保：

- [ ] 所有测试通过（`go test ./...`）
- [ ] 测试覆盖率达标（`go test -cover ./...`）
- [ ] 无数据竞争（`go test -race ./...`）
- [ ] 代码已格式化（`gofmt -s -w .` 或 `goimports -w .`）
- [ ] 通过静态检查（`go vet ./...`）
- [ ] 通过 linter（`golangci-lint run`）
- [ ] 所有导出标识符有文档注释
- [ ] 错误都已正确处理
- [ ] 没有 goroutine 泄漏
- [ ] 使用 context 控制生命周期
- [ ] 遵循命名规范
- [ ] 接口设计合理
- [ ] 代码简洁清晰，无过度设计
- [ ] 依赖已整理（`go mod tidy`）

## 13. 工具链

### 必需工具
```bash
# 安装推荐工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golang/mock/mockgen@latest
```

### 编辑器配置
推荐配置 VS Code / GoLand：
- 保存时自动格式化（goimports）
- 保存时运行 go vet
- 显示 lint 警告
- 自动导入管理

### Makefile 示例
```makefile
.PHONY: test lint fmt vet build

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

build:
	go build -v ./...

check: fmt vet lint test
```

## 14. 版本和兼容性

### Go 版本
- 使用稳定版本的 Go（当前建议 Go 1.21+）
- 在 `go.mod` 中明确指定 Go 版本
- 使用新特性前确认团队使用的 Go 版本

### API 兼容性
- 遵循语义化版本（Semantic Versioning）
- 不破坏公共 API 的向后兼容性
- 使用 `// Deprecated:` 标记废弃的 API
- 在主要版本升级时才引入破坏性变更

## 总结

遵循本规范能够：
- ✅ 提高代码质量和可维护性
- ✅ 减少 bug 和安全问题
- ✅ 提升团队协作效率
- ✅ 保持代码库的一致性
- ✅ 简化代码审查流程

**核心原则**：简单、清晰、实用。当有疑问时，优先考虑代码的可读性和可维护性。

