# rare_backend


## 项目启动命令
go run cmd/server/main.go


罕见病相关业务后端 API（Go + Gin + MySQL）。

## 环境要求

- Go 1.25+
- MySQL 8.x（已创建数据库 `rare_backend`）

## 配置与安全

敏感信息**不得**写入 Go 源码，仅通过 **`.env`** 或**环境变量**提供：

| 变量 | 必填 | 说明 |
|------|------|------|
| `MYSQL_DSN` | 是 | MySQL 连接串 |
| `JWT_SECRET` | 是 | JWT 签名密钥（长随机串） |
| `SERVER_ADDR` | 否 | 默认 `:8080` |
| `JWT_EXPIRE_HOURS` | 否 | 默认 `24` |
| `CORS_ALLOW_ORIGINS` | 否 | 逗号分隔的前端源 |
| `GIN_MODE` | 否 | 默认 `debug` |

未配置 `MYSQL_DSN` 或 `JWT_SECRET` 时，服务启动会直接失败并打印配置提示。

### 首次启动

在项目根目录：

```bash
cp .env.example .env
```

编辑 `.env`，填写本机 MySQL 密码与 JWT 密钥。`.env` 已加入 `.gitignore`，请勿提交。

也可不用 `.env`，在启动前导出环境变量：

```bash
export MYSQL_DSN='root:你的密码@tcp(127.0.0.1:3306)/rare_backend?parseTime=true'
export JWT_SECRET='你的长随机密钥'
```

### 数据库 Migration

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 管理表结构，详见 [migrations/README.md](migrations/README.md)。

```bash
make migrate-install   # 首次：安装 migrate CLI
make migrate-up        # 空库建表（43 张表）
make migrate-version   # 查看当前版本
```

`schema_reference.sql` 仅作结构参考；**可执行**的 migration 为 `000001_baseline.up.sql` 及后续增量脚本。

## 启动服务

请在**项目根目录**执行（以便加载 `./.env`）：

```bash
go run cmd/server/main.go
```

成功时示例：

```text
[config] 已加载 .env
server listening on :8080
```

健康检查：`GET http://localhost:8080/health`

## 目录结构

| 路径 | 说明 |
|------|------|
| `cmd/server/main.go` | 启动入口 |
| `internal/config` | 配置加载（`.env` + 环境变量，必填项校验） |
| `internal/router` | 路由注册 |
| `internal/middleware` | 鉴权等中间件 |
| `internal/module/*` | 业务模块（auth、community、knowledge、region、resource） |
| `internal/pkg` | 公共包（db、jwt、hash、[response](docs/api-response-contract.md)） |
| `migrations/` | 数据库 migration（[说明](migrations/README.md)） |
| `.env.example` | 配置模板（可提交 Git） |

## 模块说明

- **auth**：登录注册、用户管理（`/api/system/users` 需管理员 Token）
- **community**：病友交流区（帖子、评论）
- **knowledge**：知识库（疾病、分类、标签、文章）
- **region**：行政区划
- **resource**：医院、药品、医保、慈善、康复等资源

## 鉴权说明

- 登录 `POST /api/auth/login` 返回 JWT，载荷含 `user_id`、`role`（1 普通 / 2 专家 / 9 管理员）。
- 请求头：`Authorization: Bearer <token>`
- **管理员写接口**（`role=9`）：`/api/system/users/*`、knowledge/region/resource 的 POST/PUT/DELETE。
- **登录用户写接口**：社区发帖/评论/点赞、部分药品用户操作。
- **公开读**：各模块 GET；社区读接口支持可选 Token（`is_liked` 等）。

> 旧 Token 无 `role` 需重新登录。管理员需在数据库 `user.role = 9`。

## API 响应格式

所有 JSON 接口统一返回 `{ code, message, data }`，详见 [API 响应契约](docs/api-response-contract.md)。

## 启动顺序（main）

1. `config.Load()` — 加载 `.env` 并校验必填项  
2. `db.InitMySQL` — 连接数据库  
3. `jwt.Init` — 初始化 JWT  
4. Gin + CORS + 注册路由  
5. `r.Run` — 监听端口  





第一阶段（P0，1–2 周）
├── 启动顺序 + 配置外置（config/env）
├── JWT 校验中间件 + 管理接口鉴权/角色
├── community：post_like 字段、DeletePost Scan、去掉 user_id 硬编码/body 伪造
└── 密钥/密码移出仓库

第二阶段（P1，2–4 周）
├── 引入 middleware（日志、recovery、auth）
├── 按模块拆分 handler → service → repo（先从 auth、community 试点）
├── 统一 API 响应契约（code/message/data）
└── 数据库 migration（golang-migrate + baseline，见 migrations/）

第三阶段（P2）
├── 消除 N+1（resource 列表 + 社区点赞态批量查询）
├── 列表去掉大字段、点赞计数改为增量
└── 索引与慢查询（LIKE、JOIN）

第四阶段（P3）
├── 去 fmt.Printf、统一命名与错误处理
├── 抽公共 pagination/sqlbuilder/response
└── 补测试与 CI