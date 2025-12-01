# KubeZ 数据库迁移

## 概述

KubeZ 使用 [golang-migrate](https://github.com/golang-migrate/migrate) 进行数据库迁移管理。

## 配置

### 数据库连接

可以通过以下方式配置数据库连接：

1. **环境变量**（推荐）:
```bash
export KUBEZ_DB_URL="postgres://postgres:postgres@localhost:5432/kubez?sslmode=disable"
```

2. **命令行参数**:
```bash
kubez migrate up --db-url="postgres://postgres:postgres@localhost:5432/kubez?sslmode=disable"
```

### 迁移脚本路径

默认路径为 `./migrations`，也可以自定义：

```bash
export KUBEZ_MIGRATIONS_PATH="/path/to/migrations"
# 或
kubez migrate up --migrations-path="/path/to/migrations"
```

## 命令

### 查看迁移状态

```bash
kubez migrate status
```

输出示例：
```
当前迁移版本: 1
状态: clean (正常)
```

### 执行迁移

```bash
kubez migrate up
```

输出示例：
```
开始执行数据库迁移...
✓ 数据库迁移完成，当前版本: 1
```

### 回滚迁移

```bash
kubez migrate down
```

输出示例：
```
开始回滚数据库迁移...
✓ 数据库回滚完成，当前版本: 0
```

## 开发环境设置

### 使用 Docker Compose 启动 PostgreSQL

```bash
cd dev
docker-compose up -d postgres
```

### 创建数据库

数据库会在第一次运行迁移时自动创建表结构。如果需要手动创建数据库：

```bash
docker exec -it postgres psql -U postgres -c "CREATE DATABASE kubez;"
```

### 运行迁移

```bash
cd kubez
go build -o kubez .
./kubez migrate up
```

## 创建新的迁移

迁移文件命名规范：`{version}_{name}.{up|down}.sql`

示例：
- `000001_init.up.sql` - 初始化迁移（向上）
- `000001_init.down.sql` - 初始化迁移回滚（向下）

### 创建用户表迁移示例

**000002_create_users.up.sql**:
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    is_super_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
```

**000002_create_users.down.sql**:
```sql
DROP TABLE IF EXISTS users;
```

## 测试

### 运行测试

```bash
go test ./internal/database/... -v
```

### 测试覆盖率

```bash
go test ./internal/database/... -cover
```

当前覆盖率：**84.4%** ✓

### 运行带数据竞争检测的测试

```bash
go test ./internal/database/... -race
```

## 架构设计

### 目录结构

```
kubez/
├── cmd/
│   └── migrate.go          # 迁移命令 CLI
├── internal/
│   └── database/
│       ├── migrate.go      # 迁移器实现
│       └── migrate_test.go # 迁移测试
└── migrations/
    ├── 000001_init.up.sql
    └── 000001_init.down.sql
```

### 核心类型

**Migrator**: 数据库迁移器
- `NewMigrator(dbURL, migrationsPath)` - 创建迁移器
- `Up()` - 执行所有待执行的迁移
- `Down()` - 回滚最后一次迁移
- `Version()` - 获取当前版本和状态
- `Close()` - 关闭连接

## 故障排除

### 迁移处于 dirty 状态

如果迁移失败导致数据库处于不一致状态：

```bash
# 1. 连接到数据库
psql -U postgres -d kubez

# 2. 检查 schema_migrations 表
SELECT * FROM schema_migrations;

# 3. 手动修复 dirty 状态（谨慎操作）
UPDATE schema_migrations SET dirty = false WHERE version = <version>;
```

### 连接被拒绝

确保 PostgreSQL 正在运行：

```bash
docker ps | grep postgres
docker exec postgres pg_isready -U postgres
```

### 迁移文件未找到

确保迁移路径正确：

```bash
ls -la migrations/
```

## 最佳实践

1. ✅ **总是创建成对的 up/down 文件**
2. ✅ **在 up 迁移中使用 IF NOT EXISTS**
3. ✅ **在 down 迁移中使用 IF EXISTS**
4. ✅ **提交前测试迁移的 up 和 down**
5. ✅ **保持迁移脚本幂等性**
6. ✅ **不要修改已经应用的迁移文件**
7. ✅ **使用事务确保原子性**

## 参考

- [golang-migrate 文档](https://github.com/golang-migrate/migrate)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
