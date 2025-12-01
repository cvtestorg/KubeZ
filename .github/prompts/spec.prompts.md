# KubeZ - 技术规范文档

## Technical Specification - MVP 版本

**文档版本**: v1.0-MVP  
**创建日期**: 2025年12月1日  
**基于**: PRD v1.0-MVP  
**技术栈**: Kubernetes 1.28+, Karmada v1.8+

---

## 1. 技术架构

### 1.1 架构概览

KubeZ 采用前后端分离架构，基于 Karmada 实现多集群统一管理。

**架构分层**: 用户浏览器 → 前端层 → 后端层 → 数据层/Karmada/成员集群

### 1.2 技术选型

#### 1.2.1 前端技术栈

- **框架**: Next.js 16 with App Router
- **语言**: TypeScript 5.0+
- **UI 组件库**: shadcn/ui (基于 Radix UI)
- **样式方案**: Tailwind CSS v4
- **主题管理**: next-themes
- **图标库**: lucide-react
- **状态管理**: TanStack Query (React Query) v5 + Zustand
- **认证**: Auth.js v5 (NextAuth.js) + Keycloak
- **表单**: React Hook Form + Zod
- **图表库**: Recharts
- **HTTP 客户端**: fetch API (native)
- **实时通信**: WebSocket API

#### 1.2.2 后端技术栈

- **开发语言**: Go 1.21+
- **Web 框架**: Gin v1.10+
- **CLI 框架**: Cobra v1.8+ + Viper (配置管理)
- **数据库**:
  - 开发环境: SQLite 3.x
  - 生产环境: PostgreSQL 15+
- **ORM**: GORM v1.25+
- **认证**: 与 Keycloak 集成 (OAuth 2.0 / OIDC)
- **权限管理**: OpenFGA v1.5+ (Fine-Grained Authorization)
- **Kubernetes 客户端**: client-go v0.28+
- **Karmada 客户端**: Karmada SDK v1.8+

#### 1.2.3 多集群管理方案

- **多集群管理**: Karmada v1.8+
- **部署模式**: Host Kubernetes 集群模式
- **集群注册方式**: Push 模式
- **资源调度**: Karmada Scheduler

---

## 2. 数据模型设计

### 2.1 数据库选型

- **开发环境**: SQLite 3.x
- **生产环境**: PostgreSQL 15+
- **ORM 框架**: GORM

### 2.2 核心数据表

#### 2.2.1 用户表 (users)

- 存储系统用户信息
- 字段: id, username, password_hash, email, is_super_admin, created_at, updated_at
- 索引: username

#### 2.2.2 集群表 (clusters)

- 存储 Kubernetes 集群连接信息
- 字段: id, name, display_name, karmada_cluster_name, labels, kubeconfig, api_server, version, status, last_health_check, health_info, created_by, created_at, updated_at
- 索引: status, karmada_cluster_name

#### 2.2.3 会话表 (sessions)

- 存储用户会话信息
- 字段: id, user_id, token, expires_at, created_at
- 索引: token, expires_at

### 2.3 数据安全

- 密码存储: bcrypt 算法加密
- Kubeconfig 存储: AES-256-GCM 加密
- 加密密钥: 通过环境变量配置

---

## 3. Karmada 集成方案

### 3.1 Karmada 架构

**部署模式**: Host Kubernetes 集群模式

**核心组件**:

- karmada-apiserver: Karmada API 服务器
- karmada-controller-manager: 控制器管理器
- karmada-scheduler: 调度器
- karmada-webhook: 准入 Webhook
- etcd: Karmada 元数据存储

### 3.2 集群注册方式

**Push 模式**: KubeZ Backend 作为控制平面，通过 Karmada API 注册成员集群，Karmada 主动推送配置到成员集群

**注册流程要求**: 验证 kubeconfig → 测试连接 → 创建 Cluster 资源 → 存储凭证 → 部署 agent → 保存信息

### 3.3 多集群资源查询

**查询方式**: 通过 Karmada API 统一查询所有集群资源，或直接连接成员集群 API Server

**支持的资源类型**: Nodes, Pod, Deployment, Service, Namespace, Events

---

## 4. API 接口规范

### 4.1 API 设计原则

- RESTful 风格
- 版本管理: /api/v1
- 统一响应格式: JSON
- 标准化错误处理
- 支持分页和过滤

### 4.2 认证授权

- 认证方式: OAuth 2.0 / OpenID Connect (Keycloak)
- 前端认证: Auth.js v5 + Keycloak Provider
- 后端认证: 验证 Keycloak JWT Token
- Token 使用: `Authorization: Bearer <token>`
- 权限控制: OpenFGA (Relationship-Based Access Control)

**OpenFGA 授权模型**:

- 用户 (user)
- 组织 (organization)
- 集群 (cluster)
- 关系类型: owner, admin, editor, viewer, member
- 权限继承: organization member → cluster viewer

### 4.3 核心 API 端点

#### 4.3.1 认证接口

- POST /api/v1/auth/login
- POST /api/v1/auth/logout

#### 4.3.2 集群管理接口

- GET /api/v1/clusters
- POST /api/v1/clusters
- GET /api/v1/clusters/:id
- PUT /api/v1/clusters/:id
- DELETE /api/v1/clusters/:id
- POST /api/v1/clusters/:id/health-check

#### 4.3.3 命名空间接口

- GET /api/v1/clusters/:cluster_id/namespaces

#### 4.3.4 节点接口

- GET /api/v1/clusters/:cluster_id/nodes
- GET /api/v1/clusters/:cluster_id/nodes/:node_name

#### 4.3.5 Pod 接口

- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods
- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name
- DELETE /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name
- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/events

#### 4.3.6 日志接口

- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/logs
- WebSocket /api/v1/ws/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/logs

#### 4.3.7 工作负载接口

- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments
- GET /api/v1/clusters/:cluster_id/namespaces/:namespace/services

#### 4.3.8 监控指标接口

- GET /api/v1/clusters/:cluster_id/metrics

### 4.4 响应格式规范

**成功响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

**错误响应**:

```json
{
    "code": 40001,
    "message": "error description",
    "error": "detailed error"
}
```

**分页响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "page": 1,
        "page_size": 20,
        "items": []
    }
}
```

---

## 5. 核心服务设计

### 5.1 集群健康检查服务

**功能**: 定期检测集群健康状态

**检查周期**: 30 秒

**检查项**:

- API Server 连通性
- API 响应时间
- 集群证书有效期
- 节点状态统计
- Pod 状态统计

**结果处理**:

- 更新集群状态（online/offline/error）
- 记录健康信息（延迟、证书剩余天数）
- 更新最后检查时间

### 5.2 实时日志服务

**功能**: 提供 Pod 容器实时日志流

**通信协议**: WebSocket

**连接参数**:

- cluster_id: 集群 ID
- namespace: 命名空间
- pod_name: Pod 名称
- container: 容器名称
- token: JWT token（查询参数）

**客户端控制**:

- pause: 暂停日志流
- resume: 恢复日志流
- close: 关闭连接

**服务端消息格式**:

```json
{
    "type": "log",
    "timestamp": "2025-12-01T10:00:00Z",
    "content": "log line content"
}
```

### 5.3 监控指标采集服务

**数据源**: Kubernetes Metrics Server

**采集指标**:

- 集群级别: CPU 使用率、内存使用率、Pod/Node 数量
- 节点级别: CPU 使用率、内存使用率、Pod 数量
- Pod 级别: CPU 使用率、内存使用率、重启次数

**数据处理**:

- 实时查询（不存储历史数据）
- 按时间范围聚合（1h/24h/7d）
- 计算使用百分比

### 5.4 OpenFGA 权限服务

**功能**: 细粒度权限管理和访问控制

**授权模型定义**:

```text
model
  schema 1.1

type user

type organization
  relations
    define member: [user]
    define admin: [user]

type cluster
  relations
    define owner: [user]
    define admin: [user, organization#member]
    define editor: [user, organization#member] 
    define viewer: [user, organization#member]
    
    define can_view: viewer or editor or admin or owner
    define can_edit: editor or admin or owner
    define can_delete: admin or owner
    define can_manage: owner
```

**核心操作**:

1. **权限检查 (Check)**:
   - 输入: user, relation, object
   - 输出: allowed (true/false)
   - 示例: `Check(user:alice, can_view, cluster:prod-01)`

2. **关系写入 (Write)**:
   - 创建关系: `Write(user:alice, owner, cluster:prod-01)`
   - 删除关系: `Delete(user:alice, owner, cluster:prod-01)`

3. **列表查询 (ListObjects)**:
   - 查询用户可访问的资源
   - 示例: `ListObjects(user:alice, can_view, cluster)`

4. **关系读取 (Read)**:
   - 查询对象的所有关系
   - 示例: `Read(cluster:prod-01)` → 返回所有用户权限

**权限初始化流程**:

1. 用户注册时在 users 表创建记录
2. 首个用户自动设置为 super_admin
3. 用户添加集群时，自动调用 OpenFGA Write API 添加关系:
   `tuple: (user:{user_id}, owner, cluster:{cluster_id})`
4. 授权其他用户时，管理员调用 Write API 添加关系

**API 中间件集成**:

```text
请求流程:
1. 提取 JWT token 获取 user_id
2. 解析请求路径获取资源 (cluster_id)
3. 根据操作类型确定所需权限 (can_view/can_edit/can_delete)
4. 调用 OpenFGA Check API
5. 允许/拒绝请求
```

**性能优化**:

- 权限检查结果缓存（TTL: 60秒）
- 批量检查使用 BatchCheck API
- 对 super_admin 跳过 OpenFGA 检查

### 5.5 Kubernetes 客户端管理服务

**功能**: 管理多个集群的 Kubernetes 客户端

**客户端池**:

- 为每个集群创建独立的 clientset
- 缓存 clientset 实例，避免重复创建
- 处理客户端失效和重连

**kubeconfig 处理**:

- 从数据库读取加密的 kubeconfig
- 解密 kubeconfig
- 创建 clientset

### 5.6 CLI 管理 (Cobra)

**功能**: 提供命令行工具管理 KubeZ 后端

**主命令**: `kubez`

**子命令结构**:

```text
kubez
├── server          # 服务器管理
│   ├── start       # 启动 API 服务器
│   └── health      # 健康检查
├── migrate         # 数据库迁移
│   ├── up          # 执行迁移
│   ├── down        # 回滚迁移
│   └── status      # 迁移状态
├── config          # 配置管理
│   ├── validate    # 验证配置文件
│   ├── show        # 显示当前配置
│   └── init        # 初始化配置文件
├── user            # 用户管理
│   ├── create      # 创建用户
│   ├── list        # 列出用户
│   └── delete      # 删除用户
└── version         # 版本信息
```

**核心命令说明**:

1. **kubez server start**
   - 启动 Gin Web 服务器
   - 支持参数：`--port`, `--host`, `--config`
   - 从配置文件或环境变量读取配置（Viper）

2. **kubez migrate up**
   - 运行数据库迁移
   - 自动创建表结构
   - 支持 PostgreSQL 和 SQLite

3. **kubez config validate**
   - 验证配置文件格式
   - 检查必需字段
   - 测试数据库连接

4. **kubez user create**
   - 创建管理员用户
   - 初始化时创建首个 super_admin

**Viper 配置管理**:

- 支持配置文件：YAML、JSON、TOML
- 默认配置文件路径：`./config.yaml`, `/etc/kubez/config.yaml`
- 环境变量前缀：`KUBEZ_`
- 配置优先级：CLI 参数 > 环境变量 > 配置文件 > 默认值

**典型部署命令**:

```bash
# 初始化配置
kubez config init --output config.yaml

# 运行数据库迁移
kubez migrate up --config config.yaml

# 创建管理员用户
kubez user create --username admin --password <password> --super-admin

# 启动服务器
kubez server start --config config.yaml --port 8080
```

- 创建 REST Config
- 创建 Kubernetes Clientset

### 4.3 集群管理 API

#### GET /api/v1/clusters

获取集群列表

**查询参数**:

- `status`: 过滤状态 (online/offline/error)
- `search`: 搜索关键词
- `page`: 页码 (默认 1)
- `page_size`: 每页数量 (默认 20)

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 5,
        "items": [
            {
                "id": 1,
                "name": "prod-cluster-1",
                "display_name": "生产集群 1",
                "karmada_cluster_name": "member1",
                "labels": {"env": "prod", "region": "us-west"},
                "api_server": "https://10.0.1.100:6443",
                "version": "v1.28.3",
                "status": "online",
                "node_count": 10,
                "pod_count": 245,
                "health_info": {
                    "latency_ms": 45,
                    "cert_expiry_days": 365
                },
                "last_health_check": "2025-12-01T10:30:00Z",
                "created_at": "2025-11-01T08:00:00Z"
            }
        ]
    }
}
```

#### POST /api/v1/clusters

注册新集群

**请求体**:

```json
{
    "name": "prod-cluster-2",
    "display_name": "生产集群 2",
    "labels": {"env": "prod", "region": "eu-west"},
    "kubeconfig": "apiVersion: v1\nkind: Config\nclusters:\n..."
}
```

**响应**:

```json
{
    "code": 0,
    "message": "cluster registered successfully",
    "data": {
        "id": 2,
        "name": "prod-cluster-2",
        "karmada_cluster_name": "member2",
        "status": "online"
    }
}
```

#### GET /api/v1/clusters/:id

获取集群详情

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "name": "prod-cluster-1",
        "display_name": "生产集群 1",
        "api_server": "https://10.0.1.100:6443",
        "version": "v1.28.3",
        "status": "online",
        "node_count": 10,
        "pod_count": 245,
        "namespaces": ["default", "kube-system", "production"],
        "metrics": {
            "cpu_usage_percent": 45.2,
            "memory_usage_percent": 62.8,
            "pod_status": {
                "running": 230,
                "pending": 10,
                "failed": 5
            },
            "node_status": {
                "ready": 9,
                "not_ready": 1
            }
        },
        "health_info": {
            "latency_ms": 45,
            "cert_expiry_days": 365
        }
    }
}
```

#### PUT /api/v1/clusters/:id

更新集群信息

**请求体**:

```json
{
    "display_name": "生产集群 1 (更新)",
    "labels": {"env": "prod", "region": "us-west", "tier": "critical"}
}
```

#### DELETE /api/v1/clusters/:id

删除集群

**响应**:

```json
{
    "code": 0,
    "message": "cluster deleted successfully"
}
```

#### POST /api/v1/clusters/:id/health-check

手动触发健康检查

**响应**:

```json
{
    "code": 0,
    "message": "health check completed",
    "data": {
        "status": "online",
        "latency_ms": 42,
        "cert_expiry_days": 365,
        "api_server_reachable": true
    }
}
```

### 4.4 命名空间 API

#### GET /api/v1/clusters/:cluster_id/namespaces

获取集群的命名空间列表

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "name": "default",
                "status": "Active",
                "created_at": "2025-10-01T00:00:00Z",
                "resource_counts": {
                    "pods": 50,
                    "deployments": 10,
                    "services": 8
                }
            },
            {
                "name": "production",
                "status": "Active",
                "created_at": "2025-11-01T00:00:00Z",
                "resource_counts": {
                    "pods": 120,
                    "deployments": 25,
                    "services": 18
                }
            }
        ]
    }
}
```

### 4.5 节点 API

#### GET /api/v1/clusters/:cluster_id/nodes

获取集群节点列表

**查询参数**:

- `status`: 过滤状态 (Ready/NotReady)
- `search`: 搜索节点名称

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "name": "node-1",
                "status": "Ready",
                "roles": ["master", "control-plane"],
                "version": "v1.28.3",
                "os_image": "Ubuntu 22.04.3 LTS",
                "kernel_version": "5.15.0-91-generic",
                "container_runtime": "containerd://1.7.7",
                "cpu_capacity": "8",
                "memory_capacity": "16Gi",
                "cpu_usage_percent": 35.5,
                "memory_usage_percent": 58.3,
                "pod_count": 25,
                "pod_capacity": 110,
                "conditions": [
                    {"type": "Ready", "status": "True"},
                    {"type": "MemoryPressure", "status": "False"},
                    {"type": "DiskPressure", "status": "False"}
                ],
                "created_at": "2025-10-01T00:00:00Z"
            }
        ]
    }
}
```

#### GET /api/v1/clusters/:cluster_id/nodes/:node_name

获取节点详情

### 4.6 Pod API

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods

获取 Pod 列表

**查询参数**:

- `status`: 过滤状态 (Running/Pending/Failed/...)
- `search`: 搜索 Pod 名称
- `node`: 过滤节点

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "name": "nginx-deployment-7d64c8d5d9-abc12",
                "namespace": "production",
                "status": "Running",
                "phase": "Running",
                "node_name": "node-1",
                "pod_ip": "10.244.1.15",
                "host_ip": "10.0.1.101",
                "restart_count": 0,
                "containers": [
                    {
                        "name": "nginx",
                        "image": "nginx:1.25.3",
                        "ready": true,
                        "restart_count": 0,
                        "state": "running"
                    }
                ],
                "conditions": [
                    {"type": "Ready", "status": "True"},
                    {"type": "ContainersReady", "status": "True"}
                ],
                "created_at": "2025-11-15T10:00:00Z",
                "started_at": "2025-11-15T10:00:05Z"
            }
        ]
    }
}
```

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name

获取 Pod 详情

#### DELETE /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name

删除 Pod (实现重启功能)

**响应**:

```json
{
    "code": 0,
    "message": "pod deleted successfully"
}
```

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/events

获取 Pod 事件

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "type": "Normal",
                "reason": "Scheduled",
                "message": "Successfully assigned production/nginx-deployment-7d64c8d5d9-abc12 to node-1",
                "source": "default-scheduler",
                "count": 1,
                "first_timestamp": "2025-11-15T10:00:00Z",
                "last_timestamp": "2025-11-15T10:00:00Z"
            },
            {
                "type": "Normal",
                "reason": "Pulled",
                "message": "Container image \"nginx:1.25.3\" already present on machine",
                "source": "kubelet",
                "count": 1,
                "first_timestamp": "2025-11-15T10:00:03Z",
                "last_timestamp": "2025-11-15T10:00:03Z"
            }
        ]
    }
}
```

### 4.7 日志 API

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/logs

获取 Pod 日志（历史日志）

**查询参数**:

- `container`: 容器名称（多容器 Pod 必需）
- `tail_lines`: 尾部行数 (默认 1000)
- `since_seconds`: 最近 N 秒的日志
- `timestamps`: 是否显示时间戳 (true/false)
- `previous`: 是否查看之前容器的日志 (true/false)
- `search`: 搜索关键词

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "logs": "2025-12-01T10:30:00.123Z [INFO] Server started on port 8080\n2025-12-01T10:30:01.456Z [INFO] Connected to database\n...",
        "container": "nginx",
        "pod": "nginx-deployment-7d64c8d5d9-abc12",
        "line_count": 1000
    }
}
```

#### WebSocket /api/v1/ws/clusters/:cluster_id/namespaces/:namespace/pods/:pod_name/logs

实时日志流（WebSocket）

**连接参数**:

- `container`: 容器名称
- `token`: JWT token (查询参数)

**WebSocket 消息格式**:

```json
{
    "type": "log",
    "timestamp": "2025-12-01T10:30:05.789Z",
    "content": "[INFO] Request processed successfully"
}
```

**客户端控制消息**:

```json
{"action": "pause"}   // 暂停日志流
{"action": "resume"}  // 恢复日志流
{"action": "close"}   // 关闭连接
```

### 4.8 工作负载 API

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments

获取 Deployment 列表

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "name": "nginx-deployment",
                "namespace": "production",
                "replicas": {
                    "desired": 3,
                    "current": 3,
                    "available": 3,
                    "ready": 3
                },
                "strategy": "RollingUpdate",
                "images": ["nginx:1.25.3"],
                "labels": {"app": "nginx"},
                "selector": {"app": "nginx"},
                "created_at": "2025-11-01T00:00:00Z",
                "updated_at": "2025-11-15T10:00:00Z"
            }
        ]
    }
}
```

#### GET /api/v1/clusters/:cluster_id/namespaces/:namespace/services

获取 Service 列表

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "name": "nginx-service",
                "namespace": "production",
                "type": "ClusterIP",
                "cluster_ip": "10.96.100.50",
                "external_ips": [],
                "ports": [
                    {
                        "name": "http",
                        "protocol": "TCP",
                        "port": 80,
                        "target_port": 8080
                    }
                ],
                "selector": {"app": "nginx"},
                "session_affinity": "None",
                "created_at": "2025-11-01T00:00:00Z"
            }
        ]
    }
}
```

### 4.9 监控指标 API

#### GET /api/v1/clusters/:cluster_id/metrics

获取集群监控指标

**查询参数**:

- `time_range`: 时间范围 (1h/24h/7d)

**响应**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "cluster_id": 1,
        "time_range": "24h",
        "cpu": {
            "current_usage_percent": 45.2,
            "history": [
                {"timestamp": "2025-12-01T09:00:00Z", "value": 42.1},
                {"timestamp": "2025-12-01T10:00:00Z", "value": 45.2}
            ]
        },
        "memory": {
            "current_usage_percent": 62.8,
            "history": [
                {"timestamp": "2025-12-01T09:00:00Z", "value": 60.5},
                {"timestamp": "2025-12-01T10:00:00Z", "value": 62.8}
            ]
        },
        "pod_status": {
            "running": 230,
            "pending": 10,
            "failed": 5,
            "succeeded": 0,
            "unknown": 0
        },
        "node_status": {
            "ready": 9,
            "not_ready": 1,
            "unknown": 0
        }
    }
}
```

---

## 5. 核心服务实现

### 5.1 集群健康检查服务（代码示例）

```go
package services

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "time"
    
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

// HealthCheckService 健康检查服务
type HealthCheckService struct {
    db *gorm.DB
}

// PerformHealthCheck 执行健康检查
func (s *HealthCheckService) PerformHealthCheck(ctx context.Context, cluster *models.Cluster) (*models.HealthInfo, error) {
    healthInfo := &models.HealthInfo{}
    
    // 1. 解密 kubeconfig
    kubeconfig, err := s.decryptKubeconfig(cluster.Kubeconfig)
    if err != nil {
        healthInfo.ErrorMessage = fmt.Sprintf("failed to decrypt kubeconfig: %v", err)
        return healthInfo, err
    }
    
    // 2. 创建客户端配置
    config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
    if err != nil {
        healthInfo.ErrorMessage = fmt.Sprintf("invalid kubeconfig: %v", err)
        return healthInfo, err
    }
    
    // 3. 测试 API Server 连通性和响应时间
    start := time.Now()
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        healthInfo.ErrorMessage = fmt.Sprintf("failed to create client: %v", err)
        return healthInfo, err
    }
    
    _, err = clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
    latency := time.Since(start)
    
    if err != nil {
        healthInfo.ErrorMessage = fmt.Sprintf("API server unreachable: %v", err)
        return healthInfo, err
    }
    
    healthInfo.LatencyMs = int(latency.Milliseconds())
    
    // 4. 检查证书有效期
    certExpiryDays, err := s.checkCertExpiry(config)
    if err == nil {
        healthInfo.CertExpiryDays = certExpiryDays
    }
    
    return healthInfo, nil
}

// checkCertExpiry 检查证书有效期
func (s *HealthCheckService) checkCertExpiry(config *rest.Config) (int, error) {
    tlsConfig, err := rest.TLSConfigFor(config)
    if err != nil {
        return 0, err
    }
    
    if tlsConfig == nil || len(tlsConfig.Certificates) == 0 {
        return 0, fmt.Errorf("no certificates found")
    }
    
    cert := tlsConfig.Certificates[0]
    if len(cert.Certificate) == 0 {
        return 0, fmt.Errorf("empty certificate")
    }
    
    x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
    if err != nil {
        return 0, err
    }
    
    daysUntilExpiry := int(time.Until(x509Cert.NotAfter).Hours() / 24)
    return daysUntilExpiry, nil
}

// StartPeriodicHealthCheck 启动定期健康检查（30秒周期）
func (s *HealthCheckService) StartPeriodicHealthCheck(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.checkAllClusters(ctx)
        }
    }
}

// checkAllClusters 检查所有集群
func (s *HealthCheckService) checkAllClusters(ctx context.Context) {
    var clusters []models.Cluster
    if err := s.db.Find(&clusters).Error; err != nil {
        return
    }
    
    for _, cluster := range clusters {
        healthInfo, err := s.PerformHealthCheck(ctx, &cluster)
        
        status := "online"
        if err != nil {
            status = "error"
        }
        
        now := time.Now()
        s.db.Model(&cluster).Updates(map[string]interface{}{
            "status":            status,
            "health_info":       healthInfo,
            "last_health_check": now,
        })
    }
}
```

### 5.2 实时日志服务（代码示例）

```go
package services

import (
    "bufio"
    "context"
    "io"
    
    "github.com/gorilla/websocket"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
)

// LogService 日志服务
type LogService struct {
    clientFactory *ClusterClientFactory
}

// StreamLogs 流式传输日志（WebSocket）
func (s *LogService) StreamLogs(ctx context.Context, conn *websocket.Conn, req *LogStreamRequest) error {
    // 获取集群客户端
    clientset, err := s.clientFactory.GetClientset(req.ClusterID)
    if err != nil {
        return err
    }
    
    // 创建日志流请求
    logOptions := &corev1.PodLogOptions{
        Container:  req.Container,
        Follow:     true,
        Timestamps: true,
        TailLines:  &req.TailLines,
    }
    
    stream, err := clientset.CoreV1().Pods(req.Namespace).GetLogs(req.PodName, logOptions).Stream(ctx)
    if err != nil {
        return err
    }
    defer stream.Close()
    
    // 读取日志流并发送到 WebSocket
    reader := bufio.NewReader(stream)
    paused := false
    
    // 启动协程处理客户端控制消息
    go s.handleControlMessages(conn, &paused)
    
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            if paused {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            
            line, err := reader.ReadString('\n')
            if err != nil {
                if err == io.EOF {
                    return nil
                }
                return err
            }
            
            // 发送日志到客户端
            msg := LogMessage{
                Type:      "log",
                Timestamp: time.Now(),
                Content:   line,
            }
            
            if err := conn.WriteJSON(msg); err != nil {
                return err
            }
        }
    }
}

// handleControlMessages 处理客户端控制消息
func (s *LogService) handleControlMessages(conn *websocket.Conn, paused *bool) {
    for {
        var msg ControlMessage
        if err := conn.ReadJSON(&msg); err != nil {
            return
        }
        
        switch msg.Action {
        case "pause":
            *paused = true
        case "resume":
            *paused = false
        case "close":
            return
        }
    }
}

// GetHistoricalLogs 获取历史日志
func (s *LogService) GetHistoricalLogs(ctx context.Context, req *LogHistoryRequest) (string, error) {
    clientset, err := s.clientFactory.GetClientset(req.ClusterID)
    if err != nil {
        return "", err
    }
    
    logOptions := &corev1.PodLogOptions{
        Container:    req.Container,
        Timestamps:   true,
        TailLines:    &req.TailLines,
        SinceSeconds: req.SinceSeconds,
        Previous:     req.Previous,
    }
    
    logs := clientset.CoreV1().Pods(req.Namespace).GetLogs(req.PodName, logOptions)
    logBytes, err := logs.DoRaw(ctx)
    if err != nil {
        return "", err
    }
    
    logContent := string(logBytes)
    
    // 如果有搜索关键词，进行过滤
    if req.Search != "" {
        logContent = s.filterLogs(logContent, req.Search)
    }
    
    return logContent, nil
}

// filterLogs 过滤日志内容
func (s *LogService) filterLogs(logs string, keyword string) string {
    var filtered []string
    scanner := bufio.NewScanner(strings.NewReader(logs))
    
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, keyword) {
            filtered = append(filtered, line)
        }
    }
    
    return strings.Join(filtered, "\n")
}
```

---

---

## 6. 前端技术规范

### 6.1 项目结构

**说明**: 前端项目由开发者使用 `create-next-app` 手动初始化

```text
frontend/
├── src/
│   ├── app/              # Next.js App Router
│   │   ├── layout.tsx    # 根布局
│   │   ├── page.tsx      # 首页（集群对比视图）
│   │   ├── login/        # 登录页（由 Auth.js 处理）
│   │   ├── clusters/     # 集群相关页面
│   │   │   ├── page.tsx  # 集群列表
│   │   │   └── [id]/     # 动态路由
│   │   │       ├── page.tsx       # 集群详情
│   │   │       ├── nodes/         # 节点列表
│   │   │       ├── pods/          # Pod 列表
│   │   │       ├── logs/          # 日志查看
│   │   │       └── workloads/     # 工作负载
│   │   └── api/          # API Routes (Auth.js)
│   │       └── auth/[...nextauth]/route.ts
│   ├── components/       # 可复用组件
│   │   ├── ui/           # shadcn/ui 组件
│   │   ├── cluster-selector.tsx
│   │   ├── log-viewer.tsx
│   │   ├── metrics-chart.tsx
│   │   └── theme-provider.tsx
│   ├── lib/              # 工具库
│   │   ├── api.ts        # API 调用封装
│   │   ├── auth.ts       # Auth.js 配置
│   │   ├── utils.ts      # 工具函数
│   │   └── cn.ts         # Tailwind 类名合并
│   ├── hooks/            # 自定义 Hooks
│   ├── types/            # TypeScript 类型定义
│   └── store/            # Zustand 状态管理
├── public/
├── components.json       # shadcn/ui 配置
├── package.json
├── tsconfig.json
├── next.config.js
├── tailwind.config.ts
└── auth.config.ts        # Auth.js 配置
```

### 6.2 核心组件设计

#### 6.2.1 集群选择器 (ClusterSelector)

**功能**: 全局集群切换

**特性**:

- 显示集群名称和状态图标
- 下拉选择所有可用集群
- 切换后自动刷新页面数据
- 记住用户最后选择

**数据刷新**: 30 秒自动刷新集群列表

#### 6.2.2 实时日志查看器 (LogViewer)

**功能**: WebSocket 实时日志流

**特性**:

- 类似终端的日志显示
- 自动滚动到最新
- 暂停/恢复控制
- 容器选择（多容器 Pod）
- 日志下载功能
- 关键词搜索和高亮

**技术要点**:

- 使用原生 WebSocket API
- 虚拟滚动优化（大量日志）
- 自动重连机制

#### 6.2.3 监控图表 (MetricsChart)

**功能**: 显示集群/节点/Pod 监控指标

**图表库**: Apache ECharts 5.x

**图表类型**:

- 折线图：CPU/内存使用率趋势
- 饼图：Pod 状态分布
- 柱状图：节点资源对比

**时间范围**: 1h / 24h / 7d 切换

### 6.3 状态管理方案

**服务端数据**: TanStack Query (React Query)

**优势**:

- 自动缓存和刷新
- 请求去重
- 乐观更新
- 离线支持
- Next.js Server Components 集成

**缓存策略**:

- 集群列表：30 秒
- 节点/Pod 列表：10 秒
- 监控指标：5 秒
- 用户信息：由 Auth.js 管理

**客户端状态**: Zustand

**用途**:

- 全局 UI 状态（主题、侧边栏展开等）
- 当前选中的集群
- WebSocket 连接状态

### 6.4 路由设计 (Next.js App Router)

```text
/                       # 首页（集群对比视图）
/login                  # 登录页（Auth.js 自动处理）
/clusters               # 集群列表
/clusters/[id]          # 集群详情
/clusters/[id]/nodes    # 节点列表
/clusters/[id]/pods     # Pod 列表
/clusters/[id]/logs     # 日志查看
/clusters/[id]/workloads # 工作负载
/api/auth/[...nextauth] # Auth.js API 路由
```

### 6.5 UI 和样式方案

**组件库**: shadcn/ui

**特点**:

- 基于 Radix UI 的无样式组件
- 完全可定制（复制到项目中，非 npm 依赖）
- TypeScript 原生支持
- Tailwind CSS 样式

**样式方案**:

- **基础样式**: Tailwind CSS v4
- **主题**: next-themes（支持亮色/暗色模式）
- **图标**: lucide-react
- **响应式**: 移动端优先设计

### 6.6 认证集成 (Auth.js + Keycloak)

**Auth.js 配置** (`src/lib/auth.ts`):

- Provider: KeycloakProvider
- Session 策略: JWT
- Callbacks: jwt、session（添加自定义字段）

**Keycloak 配置要求**:

- Realm: kubez 或自定义
- Client ID: kubez-frontend
- Client Type: Public（Next.js 为公开客户端）
- Valid Redirect URIs: `http://localhost:3000/api/auth/callback/keycloak`
- Web Origins: `http://localhost:3000`

**认证流程**:

1. 用户点击登录 → 重定向到 Keycloak 登录页
2. Keycloak 验证 → 返回 Authorization Code
3. Auth.js 交换 code 获取 access_token 和 refresh_token
4. 创建 Next.js session（存储在加密 cookie）
5. 后续请求携带 access_token 访问 Backend API

**Token 传递**:

- 前端通过 `useSession()` 获取 token
- API 调用时添加 `Authorization: Bearer <token>` 头
- 后端验证 Keycloak 签发的 JWT

**用户信息同步**:

- Keycloak 用户首次登录时，Backend 自动创建 users 表记录
- 使用 Keycloak user_id 作为唯一标识
- 权限由 OpenFGA 管理

---

## 7. 部署方案

### 7.1 容器化

#### 7.1.1 后端 Dockerfile

**基础镜像**: golang:1.21-alpine

**构建阶段**:

1. 复制 go.mod 和 go.sum
2. 下载依赖（包括 Cobra、Viper 等）
3. 复制源代码
4. 编译二进制文件（使用 CGO_ENABLED=0）

**运行镜像**: alpine:3.18

**配置**:

- 非 root 用户运行
- 暴露 8080 端口
- 挂载 kubeconfig 目录
- 支持 Cobra 命令行参数

**CLI 命令**:

```bash
kubez server start          # 启动 API 服务器
kubez version               # 显示版本信息
kubez config validate       # 验证配置文件
kubez migrate up            # 运行数据库迁移
```

#### 7.1.2 前端 Dockerfile

**说明**: 前端项目由开发者手动初始化，使用 Next.js 16 standalone 输出模式

**构建镜像**: node:20-alpine

**运行镜像**: node:20-alpine（Next.js 需要 Node.js 运行时）

**配置**:

- 启用 Next.js standalone 输出模式（next.config.js）
- 仅复制 .next/standalone 和 public 目录
- 暴露 3000 端口
- 非 root 用户运行

**环境变量**:

- NEXTAUTH_URL: 应用访问地址
- NEXTAUTH_SECRET: Auth.js 密钥
- KEYCLOAK_CLIENT_ID: Keycloak 客户端 ID
- KEYCLOAK_CLIENT_SECRET: Keycloak 客户端密钥
- KEYCLOAK_ISSUER: Keycloak Issuer URL

### 7.2 Docker Compose 部署

**说明**: OpenFGA 和 Keycloak 为外部独立部署的服务，KubeZ 只需配置连接信息

**服务组成**:

- postgres: PostgreSQL 15（KubeZ 数据库）
- backend: KubeZ API Server
- frontend: Next.js 16 应用

**外部依赖**:

- Keycloak: 外部认证服务（需提前部署）
- OpenFGA: 外部权限服务（需提前部署）
- Karmada: 外部多集群管理服务（需提前部署）

**网络**: 内部桥接网络

**存储卷**:

- postgres_data: 数据库数据
- kubeconfig: 集群配置文件

**端口映射**:

- frontend: 3000 (Next.js)
- backend: 8080 (API)

**环境变量配置**:

- `KEYCLOAK_URL`: 外部 Keycloak 服务地址
- `OPENFGA_API_URL`: 外部 OpenFGA 服务地址
- `KARMADA_KUBECONFIG`: Karmada kubeconfig 文件路径

### 7.3 Kubernetes 部署

#### 7.3.1 部署资源

**Backend**:

- Deployment: 2 副本
- Service: ClusterIP
- ConfigMap: 应用配置
- Secret: 敏感信息（数据库密码、JWT 密钥）

**Frontend**:

- Deployment: 2 副本
- Service: ClusterIP

**Database**:

- StatefulSet: PostgreSQL
- PersistentVolumeClaim: 数据持久化
- Service: Headless Service

**Ingress**:

- 统一入口
- HTTPS (cert-manager)
- 路径路由: /api → backend, / → frontend

#### 7.3.2 资源限制

**Backend**:

- requests: CPU 200m, Memory 256Mi
- limits: CPU 1000m, Memory 1Gi

**Frontend**:

- requests: CPU 100m, Memory 128Mi
- limits: CPU 500m, Memory 512Mi

**Database**:

- requests: CPU 500m, Memory 512Mi
- limits: CPU 2000m, Memory 2Gi

### 7.4 Karmada 部署

**安装方式**: Helm Chart

**命名空间**: karmada-system

**部署模式**: Host 模式（复用现有集群）

**配置存储**:

- 获取 karmada-kubeconfig
- 创建 Secret 存储到 kubez 命名空间
- Backend 挂载 Secret

---

## 8. 安全规范

### 8.1 认证安全

**密码策略**:

- 最小长度：8 位
- 包含大小写字母、数字
- 使用 bcrypt 加密（cost 12）

**JWT Token**:

- 签名算法：HS256
- 密钥长度：≥32 字节
- 有效期：7 天
- 刷新机制：滑动窗口

### 8.2 数据加密

**Kubeconfig 加密**:

- 算法：AES-256-GCM
- 密钥管理：环境变量注入
- 密钥轮换：支持密钥版本管理

**传输加密**:

- HTTPS/TLS 1.2+
- WebSocket Secure (wss://)

### 8.3 权限控制

**OpenFGA 授权模型**:

```text
model
  schema 1.1

type user

type organization
  relations
    define member: [user]
    define admin: [user]

type cluster
  relations
    define owner: [user]
    define admin: [user, organization#member]
    define editor: [user, organization#member]
    define viewer: [user, organization#member]
    
    define can_view: viewer or editor or admin or owner
    define can_edit: editor or admin or owner
    define can_delete: admin or owner
    define can_manage: owner
```

**权限矩阵**:

| 操作 | viewer | editor | admin | owner |
| ---- | ------ | ------ | ----- | ----- |
| 查看集群列表 | ✓ | ✓ | ✓ | ✓ |
| 查看集群详情 | ✓ | ✓ | ✓ | ✓ |
| 查看资源（Pods/Nodes） | ✓ | ✓ | ✓ | ✓ |
| 查看日志 | ✓ | ✓ | ✓ | ✓ |
| 重启 Pod | ✗ | ✓ | ✓ | ✓ |
| 编辑资源 | ✗ | ✓ | ✓ | ✓ |
| 添加集群 | ✗ | ✗ | ✓ | ✓ |
| 删除集群 | ✗ | ✗ | ✓ | ✓ |
| 管理用户权限 | ✗ | ✗ | ✗ | ✓ |

**API 权限检查**:

- 所有 API 需要 JWT 认证
- 每个请求通过 OpenFGA Check API 验证权限
- 示例：`user:{user_id}` 是否有 `can_view` 关系到 `cluster:{cluster_id}`

**OpenFGA 集成**:

- **部署方式**: 独立容器或使用 OpenFGA Cloud
- **SDK**: openfga/go-sdk v0.3+
- **存储**: PostgreSQL（共享或独立数据库）
- **API 端点**: gRPC (默认 8081) 或 HTTP (默认 8080)

**权限初始化**:

- 系统初始化时创建默认 admin 用户
- 用户添加集群时自动设置为该集群的 owner
- 支持通过 API 授予其他用户权限

### 8.4 安全加固

**后端**:

- 限流：API rate limiting
- 防护：CORS 配置
- 验证：输入参数校验
- 日志：审计日志记录

**数据库**:

- 连接加密：SSL/TLS
- 最小权限：专用数据库用户
- 备份：定期备份

**Kubernetes**:

- RBAC：最小权限 ServiceAccount
- Network Policy：网络隔离
- Pod Security：Security Context

---

## 9. 监控和日志规范

### 9.1 应用监控

**监控方案**: Prometheus + Grafana

**监控指标**:

- 应用指标：请求量、响应时间、错误率
- 系统指标：CPU、内存、网络
- 业务指标：集群数量、Pod 数量、用户数

**告警规则**:

- API 错误率 > 5%
- 响应时间 > 3s
- 内存使用 > 80%

### 9.2 日志规范

**日志级别**:

- ERROR: 错误信息
- WARN: 警告信息
- INFO: 关键操作
- DEBUG: 调试信息（仅开发环境）

**日志格式**: JSON 结构化日志

**日志字段**:

- timestamp: 时间戳
- level: 日志级别
- message: 日志消息
- user_id: 用户 ID（如适用）
- cluster_id: 集群 ID（如适用）
- request_id: 请求追踪 ID
- error: 错误堆栈（如有）

**日志收集方案**: Fluent Bit → Loki → Grafana

**实现说明**:

- **采集**: Fluent Bit DaemonSet 采集容器日志
- **存储**: Loki 存储和索引日志
- **查询**: Grafana 统一查询（LogQL）
- **保留**: 7 天热存储，30 天冷存储
- **标签**: 自动添加 namespace、pod、container 标签

### 9.3 健康检查

**Liveness Probe**: /health

**Readiness Probe**: /ready

**检查频率**: 10 秒

---

## 10. 性能优化规范

### 10.1 后端优化

**数据库**:

- 连接池：MaxOpenConns=25, MaxIdleConns=5
- 索引优化：为常用查询字段添加索引
- 慢查询监控：记录 > 1s 的查询

**API**:

- 分页：默认 20 条/页，最大 100 条/页
- 缓存：使用 Go 内存缓存（go-cache）缓存集群元数据
- 并发：goroutine 并发查询多集群
- 限流：Gin 中间件 + 内存存储，每用户 100 请求/分钟

**Kubernetes 客户端**:

- 客户端池：复用 clientset
- 超时控制：API 调用 30 秒超时
- 重试机制：失败自动重试 3 次

### 10.2 前端优化

**Next.js 优化特性**:

- Server Components：减少客户端 JavaScript
- 自动代码分割：按路由自动分割
- Image 优化：next/image 自动优化图片
- Font 优化：next/font 优化字体加载

**代码优化**:

- Tree Shaking：移除未使用代码
- 压缩：生产构建自动压缩
- Dynamic Import：按需加载组件

**渲染优化**:

- 虚拟滚动：大列表（> 100 项）使用 @tanstack/react-virtual
- 防抖节流：搜索输入防抖 300ms
- Streaming SSR：渐进式渲染

**缓存优化**:

- Next.js 缓存：静态资源自动缓存
- TanStack Query：智能数据缓存
- CDN 缓存：使用 Cache-Control 头

### 10.3 网络优化

**CDN**: 静态资源使用 CDN

**HTTP/2**: 开启 HTTP/2 多路复用

**压缩**: 启用 Gzip/Brotli 压缩

**Keep-Alive**: 保持连接复用

---

## 11. 测试规范

### 11.1 后端测试

**单元测试**:

- 框架：Go testing + testify
- 覆盖率：≥80%
- Mock：gomock

**集成测试**:

- 数据库：使用 SQLite 内存数据库
- API：httptest 模拟 HTTP 请求
- Kubernetes：使用 fake clientset

### 11.2 前端测试

**单元测试**:

- 框架：Vitest（Next.js 官方推荐）
- 覆盖率：≥70%
- 组件测试：React Testing Library
- Mock：MSW (Mock Service Worker) 模拟 API

**E2E 测试**:

- 框架：Playwright（Next.js 官方推荐）
- 关键流程：登录（Keycloak）、集群管理、日志查看
- 环境：独立测试环境

### 11.3 CI/CD 流程

**工具**: GitHub Actions

**触发条件**:

- Push to main: 运行完整 CI/CD
- Pull Request: 运行测试和代码检查
- Tag (v*): 构建发布版本

**CI 流程**:

1. **代码检查**
   - Go: golangci-lint, go vet
   - TypeScript: ESLint, TypeScript compiler
   - Markdown: markdownlint

2. **单元测试**
   - Backend: go test -race -cover
   - Frontend: vitest run --coverage

3. **构建验证**
   - Backend: go build
   - Frontend: next build

4. **集成测试**
   - 启动测试数据库
   - 运行 API 集成测试

5. **E2E 测试**（仅 main 分支）
   - Docker Compose 启动环境
   - Playwright 运行 E2E 测试

**CD 流程**:

1. **构建镜像**
   - 多阶段 Dockerfile 构建
   - 使用 GitHub 缓存加速
   - 镜像标签：commit SHA + latest

2. **推送镜像**
   - GitHub Container Registry (ghcr.io)
   - 或 Docker Hub / Harbor

3. **部署**（可选）
   - 开发环境：自动部署到 dev 集群
   - 生产环境：手动触发或 GitOps (ArgoCD)

**镜像仓库**: GitHub Container Registry (ghcr.io)

**缓存策略**:

- Go 依赖缓存：~/.cache/go-build
- Node 依赖缓存：~/.npm
- Docker 层缓存：GitHub Actions cache

### 11.4 开发环境配置

**本地开发工具**:

- Go 1.21+
- Node.js 20+
- Docker Desktop
- VS Code（推荐）

**VS Code 扩展推荐**:

- Go (golang.go)
- ESLint
- Prettier
- Tailwind CSS IntelliSense
- GitLens
- Docker

**启动步骤**:

1. **克隆仓库**

   ```bash
   git clone https://github.com/your-org/kubez.git
   cd kubez
   ```

2. **配置外部服务连接**（指向开发环境的 Keycloak、OpenFGA、Karmada）

   ```bash
   cp .env.example .env
   # 编辑 .env 配置外部服务地址
   ```

3. **启动后端**（热重载）

   ```bash
   cd backend
   go mod download
   air  # 使用 air 实现热重载
   ```

4. **启动前端**（热重载）

   ```bash
   cd frontend
   npm install
   npm run dev  # Next.js 自带热重载
   ```

5. **启动数据库**

   ```bash
   docker-compose up -d postgres
   ```

6. **运行数据库迁移**

   ```bash
   cd backend
   kubez migrate up
   ```

**Mock 数据**:

- 提供种子脚本：scripts/seed.sql
- 创建测试集群、测试用户
- 使用 `kubez user create` 创建本地管理员

**调试配置**:

- Backend: VS Code launch.json 配置 Delve 调试器
- Frontend: Chrome DevTools + React DevTools

**热重载**:

- Backend: 使用 air（go install github.com/cosmtrek/air@latest）
- Frontend: Next.js 自带热重载

**环境隔离**:

- 使用 .env.local 覆盖本地配置
- .env.local 不提交到 Git

---

## 12. 运维规范

### 12.1 部署流程

**前置条件**（需提前部署）:

1. Keycloak 认证服务（配置好 Realm 和 Client）
2. OpenFGA 权限服务（创建 Store 和授权模型）
3. Karmada 多集群管理平台
4. PostgreSQL 数据库

**KubeZ 部署步骤**:

1. 准备 Kubernetes 集群环境
2. 配置外部服务连接信息（Keycloak、OpenFGA、Karmada）
3. 创建 KubeZ 数据库和用户
4. 部署 Backend（运行数据库迁移）
5. 配置 Backend 环境变量（连接外部服务）
6. 部署 Frontend（配置 Auth.js 环境变量）
7. 配置 Ingress
8. 验证功能（登录、集群管理、权限）
9. 初始化首个管理员用户

### 12.2 备份策略

**数据库备份**:

- 频率：每天
- 保留：30 天
- 存储：对象存储（S3/OSS）

**配置备份**:

- Kubernetes 资源清单
- 应用配置文件

### 12.3 故障恢复

**RTO** (恢复时间目标): < 1 小时

**RPO** (恢复点目标): < 24 小时

**灾备方案**:

- 多副本部署
- 数据库主从复制
- 跨区域备份

---

## 附录

### A. 环境变量清单

```bash
# Backend
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kubez
DB_USER=kubez
DB_PASSWORD=<password>
KARMADA_KUBECONFIG=/path/to/karmada-config
ENCRYPTION_KEY=<32-byte-key>
SERVER_PORT=8080
LOG_LEVEL=info

# Keycloak 集成
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=kubez
KEYCLOAK_CLIENT_ID=kubez-backend
KEYCLOAK_CLIENT_SECRET=<client-secret>

# OpenFGA
OPENFGA_API_URL=http://openfga:8080
OPENFGA_STORE_ID=<store-id>
OPENFGA_MODEL_ID=<model-id>
# 或使用 OpenFGA Cloud
# OPENFGA_API_URL=https://api.fga.dev
# OPENFGA_API_TOKEN=<api-token>

# Frontend (Next.js)
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=<32-byte-secret>
KEYCLOAK_CLIENT_ID=kubez-frontend
KEYCLOAK_CLIENT_SECRET=<client-secret>
KEYCLOAK_ISSUER=http://keycloak:8080/realms/kubez
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### B. API 错误码表

| 错误码范围 | 说明 |
| --------- | ---- |
| 0 | 成功 |
| 40001-40099 | 认证错误（登录失败、token 无效等） |
| 40101-40199 | 参数验证错误 |
| 40301-40399 | 权限错误（无权限操作） |
| 40401-40499 | 资源不存在 |
| 50001-50099 | 服务器内部错误 |
| 50101-50199 | 数据库错误 |
| 50201-50299 | Kubernetes API 错误 |
| 50301-50399 | Karmada API 错误 |

### C. 端口分配

| 组件 | 端口 | 说明 |
| ---- | ---- | ---- |
| Backend API | 8080 | HTTP API |
| Frontend (Next.js) | 3000 | Next.js 开发服务器 |
| Frontend (生产) | 80/443 | 通过 Nginx/Ingress 代理 |
| PostgreSQL | 5432 | 数据库 |
| Karmada API | 5443 | Karmada API Server |
| Keycloak | 8180 | Keycloak 管理界面 |
| OpenFGA HTTP | 8082 | OpenFGA HTTP API |
| OpenFGA gRPC | 8081 | OpenFGA gRPC API |
| OpenFGA Playground | 3001 | OpenFGA UI (开发环境) |

### D. 技术决策记录

**为什么选择 Karmada？**

- 原生 Kubernetes 多集群管理
- 支持 Push/Pull 模式
- 活跃的社区和完善的文档
- 与 Kubernetes API 兼容

**为什么选择 OpenFGA？**

- 细粒度权限控制（Fine-Grained Authorization）
- 基于关系的访问控制模型（Relationship-Based Access Control）
- Google Zanzibar 论文实现，经过大规模验证
- 支持复杂的权限继承和委托
- 云原生设计，易于扩展
- 开源且有 CNCF 支持

**为什么选择 Next.js 16？**

- 服务端渲染（SSR）和静态生成（SSG）支持
- App Router：基于 React Server Components 的现代路由
- 自动代码分割和优化
- 内置 API Routes，方便集成 Auth.js
- Vercel 团队维护，生态成熟

**为什么选择 Auth.js + Keycloak？**

- Auth.js：Next.js 官方推荐的认证库，与 Next.js 深度集成
- Keycloak：企业级开源身份管理，支持 OAuth 2.0/OIDC/SAML
- 统一认证：可扩展到多个应用的 SSO
- 成熟稳定：Red Hat 支持，广泛应用于企业环境
- 灵活配置：支持多种认证方式和身份提供商

**为什么选择 shadcn/ui？**

- 可定制性强：代码复制到项目中，完全可控
- 无锁定风险：不是 npm 依赖，没有版本兼容问题
- 基于 Radix UI：无障碍性好，组件质量高
- Tailwind CSS：与 Next.js 生态完美配合
- TypeScript 原生支持

**为什么使用 Cobra + Viper？**

- Cobra：Go 社区标准 CLI 框架（kubectl、docker 都在用）
- 子命令支持：清晰的命令结构
- Viper：强大的配置管理，支持多种格式
- 环境变量集成：12-Factor App 最佳实践

**为什么选择 Go + Next.js？**

- Go: 高性能、并发支持好、Kubernetes 客户端成熟
- Next.js: 现代化全栈框架、SEO 友好、开发体验好

**为什么使用 PostgreSQL？**

- 开源、稳定、性能好
- 支持 JSONB 类型（灵活存储标签和元数据）
- 完善的备份和恢复工具
- KubeZ 独立数据库，不与外部服务共享

**为什么 OpenFGA 和 Keycloak 外部部署？**

- **职责分离**：认证和权限是基础设施，应独立于业务系统
- **复用性**：可为多个应用提供统一的认证和权限服务
- **可靠性**：独立部署提高可用性，避免单点故障
- **可扩展性**：独立扩展认证和权限服务，不影响业务系统
- **企业标准**：符合企业级架构的最佳实践

**为什么选择内存缓存而非 Redis？**

- **简化架构**：MVP 阶段减少外部依赖
- **性能足够**：集群元数据量小，内存缓存够用
- **运维成本**：减少一个服务的部署和维护
- **后续升级**：如有需要，可平滑升级到 Redis
- **使用方案**：go-cache 或 bigcache（Go 原生库）

**为什么选择 Karmada API 查询？**

- **统一入口**：通过 Karmada 统一查询所有集群
- **简化逻辑**：无需管理多个集群连接
- **资源聚合**：天然支持跨集群资源聚合
- **权限控制**：Karmada 层面统一权限管理
- **后续扩展**：支持资源调度和分发能力

**为什么使用 Gin 中间件限流？**

- **轻量级**：无需额外服务，直接集成到 API 中
- **足够用**：MVP 阶段流量不大，内存存储足够
- **实现简单**：Gin 生态有成熟的限流中间件
- **灵活配置**：支持按用户、IP、路径不同限流策略

#### 日志收集方案：Fluent Bit + Loki

- **轻量级**：Fluent Bit 资源占用小（相比 Fluentd）
- **生态一致**：Loki 与 Prometheus 同属 Grafana Labs，集成好
- **查询简单**：LogQL 类似 PromQL，学习成本低
- **成本低**：Loki 存储成本远低于 Elasticsearch
- **够用**：MVP 阶段不需要复杂的日志分析

#### CI/CD 选择：GitHub Actions

- **集成度高**：与 GitHub 仓库无缝集成
- **免费额度**：公开仓库免费，私有仓库有免费额度
- **生态成熟**：大量现成的 Actions 可用
- **配置简单**：YAML 配置，易于维护
- **足够强大**：支持矩阵构建、缓存、并行等高级特性

#### 开发环境：Docker Compose

- **快速启动**：一条命令启动所有服务
- **环境一致**：开发、测试环境保持一致
- **易于调试**：支持热重载和本地调试
- **低门槛**：不需要本地 Kubernetes 集群

#### WebSocket 连接管理策略

- **连接限制**：每用户最多 10 个并发 WebSocket 连接
- **心跳机制**：30 秒心跳，3 次失败断开
- **自动重连**：前端自动重连，指数退避（1s, 2s, 4s, 8s, 16s）
- **负载均衡**：使用 Sticky Session（Session Affinity）

#### 备份策略

- **数据库备份**：pg_dump 每天全量备份，保留 30 天
- **备份存储**：对象存储（S3 兼容）
- **恢复测试**：每月一次恢复演练
- **RTO/RPO**：RTO < 1 小时，RPO < 24 小时
- **不备份**：OpenFGA 和 Keycloak 数据由外部服务自行备份

#### 高可用策略（MVP 后）

- **Backend**：HPA 自动扩展，2-10 副本，基于 CPU 70%
- **Frontend**：HPA 自动扩展，2-5 副本，基于 CPU 60%
- **Database**：主从复制 + PgBouncer 连接池
- **跨区域**：MVP 阶段单区域，后续考虑多区域

---

### E. MVP 技术范围说明

**包含**:

- ✅ 基础认证（Keycloak OAuth 2.0）
- ✅ 基础权限（OpenFGA RBAC）
- ✅ 多集群管理（通过 Karmada）
- ✅ 资源查看（Node、Pod、Deployment、Service）
- ✅ 实时日志（WebSocket）
- ✅ 基础监控（Prometheus + Grafana）
- ✅ 内存缓存（go-cache）
- ✅ API 限流（Gin 中间件）

**不包含（后续版本）**:

- ❌ Redis 分布式缓存
- ❌ 消息队列（RabbitMQ/Kafka）
- ❌ 服务网格（Istio/Linkerd）
- ❌ 复杂日志分析（Elasticsearch）
- ❌ 链路追踪（Jaeger/Zipkin）
- ❌ 多租户隔离
- ❌ 资源编辑和创建
- ❌ 告警和通知系统
- ❌ 审计日志详细分析
- ❌ 跨区域高可用

---

**文档版本**: v1.0-MVP  
**最后更新**: 2025年12月1日  
**状态**: 已确认
