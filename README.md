# YouMind Backend v2

基于 **领域驱动设计（DDD）** 与 **GORM** 的 Go 后端重构版，与原有 `go-backend` API 行为兼容。

## 目录结构

```
go-backend-v2/
├── cmd/server/          # 入口
├── internal/
│   ├── config/          # 配置
│   ├── domain/          # 领域层：实体 + 仓储接口
│   │   ├── user/
│   │   ├── project/     # Project, Message, Resource
│   │   └── skill/
│   ├── application/     # 应用层：用例服务
│   │   ├── auth/
│   │   ├── project/
│   │   ├── chat/
│   │   ├── skill/
│   │   └── user/
│   ├── infrastructure/  # 基础设施：GORM 仓储 + AI 客户端
│   │   ├── persistence/
│   │   └── ai/
│   └── interfaces/      # 接口层：HTTP
│       └── http/
└── go.mod
```

## 运行

**所有配置以 `.env` 为准**，无代码内默认值（端口未设置时用 `8080` 以便启动）。

```bash
# 依赖
go mod tidy

# 复制并编辑 .env（必须设置 DATABASE_URL、JWT_SECRET）
cp .env.example .env

# 启动
go run ./cmd/server
```

## 数据库与迁移

- 启动时会**自动执行迁移**：检查 `schema_migrations` 表，将尚未执行的 `.up.sql` 按版本号顺序执行并记录，实现变更同步落地到库。
- 迁移文件位于 `internal/infrastructure/persistence/migrations/`，命名格式：`001_名称.up.sql` / `001_名称.down.sql`。新增迁移时增加新编号即可，启动时自动应用。

## API

与 `go-backend` 一致：`/health`、`/api/auth/*`、`/api/projects/*`、`/api/chat`、`/api/skills/*`、`/api/models`、`/api/user/*`。
