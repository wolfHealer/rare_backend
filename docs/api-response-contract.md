# API 响应契约规范

> 版本：v1.0  
> 适用范围：`rare_backend` 全部 HTTP JSON API（含 `/health`、`/api/*`）  
> 实现位置：`internal/pkg/response`

---

## 1. 总体原则

1. 所有业务 JSON 响应使用统一 envelope：`code` + `message` + `data`。
2. **HTTP 状态码**与 body 内 **`code` 保持一致**。
3. 成功与失败均走同一结构，客户端只需解析一套字段。
4. 业务数据只出现在 `data` 中，不在顶层散落字段。
5. 内部错误细节不得通过 `message` 暴露给客户端（仅写服务端日志）。

---

## 2. 响应结构

### 2.1 标准 Envelope

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `code` | int | 是 | 业务状态码，与 HTTP 状态码一致 |
| `message` | string | 是 | 人类可读说明；读接口成功默认 `success` |
| `data` | any | 是 | 业务载荷；无数据时统一为 `null` |

**硬性约定：**

- 字段名固定为 `message`，禁止使用 `msg`、`error`、`errMsg` 等别名。
- `data` **始终存在**：有数据则对象/数组/标量；无数据则为 `null`，**不得省略**。
- 响应 `Content-Type`：`application/json; charset=utf-8`。

### 2.2 成功响应

#### 2.2.1 普通成功（查询、更新、删除等）

```json
{
  "code": 200,
  "message": "success",
  "data": { "id": 1, "name": "示例" }
}
```

#### 2.2.2 无业务数据的成功

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

#### 2.2.3 写操作成功（可选业务文案）

创建/更新/删除允许使用业务向 `message`，`code` 仍为 **200**，HTTP 亦为 **200**：

```json
{
  "code": 200,
  "message": "创建成功",
  "data": { "id": 123 }
}
```

> 本项目**不采用** HTTP 201 / body `code: 201` 表示创建成功。

#### 2.2.4 分页列表（`data` 内结构）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "pageSize": 10
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `list` | array | 当前页数据，空时为 `[]` |
| `total` | int64 | 总条数 |
| `page` | int | 当前页，从 1 开始 |
| `pageSize` | int | 每页条数 |

**约定：**

- 列表键名统一为 `list`，不使用 `records`。
- 分页键名统一为 `page` / `pageSize`，不使用 `limit`、`size`。
- Query 参数在过渡期可兼容旧命名，**响应体**必须统一。
- 不需要分页的列表（选项、树等）：`data` 可直接为数组，或 `{ "list": [...] }`。

#### 2.2.5 分层/分组数据

分组结构放在 `data` 内，不提升到顶层。示例（医保政策列表）：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "query_mode": "region_disease",
    "selected_filters": { "disease_id": 1, "province_code": "320000" },
    "tips": "已为您展示...",
    "policies": {
      "national": [],
      "province": [],
      "city": []
    },
    "total": 10,
    "page": 1,
    "pageSize": 10
  }
}
```

---

## 3. HTTP 状态码与 `code` 映射

| HTTP | body `code` | 典型场景 |
|------|-------------|----------|
| 200 | 200 | 成功 |
| 400 | 400 | 参数错误、校验失败 |
| 401 | 401 | 未登录、Token 缺失/无效/过期 |
| 403 | 403 | 已登录但权限不足 |
| 404 | 404 | 资源不存在 |
| 500 | 500 | 服务器内部错误 |

**规则：**

- `HTTP 状态码 === body.code`。
- 禁止 HTTP 200 但 body `code` 为 4xx/5xx。
- 禁止 HTTP 5xx 但 body `code` 为 200。

---

## 4. 错误响应

### 4.1 标准错误

```json
{
  "code": 400,
  "message": "参数错误",
  "data": null
}
```

### 4.2 `message` 规范

| 类型 | 要求 | 示例 |
|------|------|------|
| 参数绑定失败 | 简短明确 | `参数错误` |
| 资源不存在 | 简短明确 | `帖子不存在` |
| 权限 | 不泄露资源是否存在 | `无权限` / `权限不足` |
| 服务器错误 | 固定文案，不含 SQL/堆栈 | `服务器错误` |

**禁止向客户端返回：**

```json
{ "message": "查询失败: Error 1064: ..." }
```

### 4.3 鉴权中间件

`internal/middleware/auth.go` 统一格式：

```json
{
  "code": 401,
  "message": "未登录或 Token 缺失",
  "data": null
}
```

```json
{
  "code": 403,
  "message": "权限不足",
  "data": null
}
```

---

## 5. 典型场景示例

### 5.1 登录成功

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": 1,
    "role": 1,
    "nickname": "用户",
    "phone": "138****",
    "avatar": "",
    "token": "eyJ...",
    "expires_in": 86400
  }
}
```

### 5.2 注册成功

```json
{
  "code": 200,
  "message": "注册成功",
  "data": null
}
```

### 5.3 健康检查

`GET /health`

```json
{
  "code": 200,
  "message": "success",
  "data": { "status": "ok" }
}
```

---

## 6. 服务端实现约定

### 6.1 统一出口

```
internal/pkg/response/
  response.go    # Body、OK、OKMessage、Fail、Page 等
```

| 函数 | 用途 |
|------|------|
| `response.OK(c, data)` | 成功，`message: success` |
| `response.OKMessage(c, msg, data)` | 成功，自定义 message |
| `response.Page(c, list, total, page, pageSize)` | 分页成功 |
| `response.BadRequest(c, msg)` | 400 |
| `response.Unauthorized(c, msg)` | 401 |
| `response.Forbidden(c, msg)` | 403 |
| `response.NotFound(c, msg)` | 404 |
| `response.InternalError(c, msg)` | 500 |

**handler、middleware 禁止直接 `c.JSON(gin.H{...})`**，须通过上述函数或模块薄封装。

### 6.2 模块层职责

```
handler  →  module/response.go（薄封装）  →  pkg/response
```

各模块 `response.go` 提供：

- `respondOK` / `respondOKMessage` / `respondPage`
- `respondBadRequest` / `respondNotFound` / …
- `respondServiceError`：domain 错误 → 标准 HTTP + message

### 6.3 新增接口 checklist

- [ ] 成功/失败均含 `code`、`message`、`data`
- [ ] 无 `msg` 字段
- [ ] HTTP 与 body `code` 一致
- [ ] 500 的 `message` 无内部错误泄露
- [ ] 分页使用 `list` / `total` / `page` / `pageSize`
- [ ] 空列表为 `[]`，非 `null`

---

## 7. 客户端解析约定

```typescript
interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T | null;
}

// 成功：code === 200
// 失败：code !== 200，展示 message
// data === null 表示无业务载荷，不等于失败
// 列表空数据：data.list === [] && data.total === 0
```

---

## 8. 破坏性变更（v1 迁移）

若前端仍对接旧版接口，需关注：

| 变更 | 旧 | 新 |
|------|----|----|
| 社区帖子列表 | `data.records` | `data.list` |
| 社区帖子分页 | `data.limit` | `data.pageSize` |
| 社区评论分页 | `data.size` | `data.pageSize` |
| 无数据成功/错误 | 可能省略 `data` | 始终 `"data": null` |
| 健康检查 | `{ "status": "ok" }` | 标准 envelope |
| 创建类接口 | 部分 HTTP 201 | 统一 HTTP 200 |
| 疾病搜索等 | `"msg"` 字段 | `"message"` |

---

## 9. 变更记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-05 | 落地 `internal/pkg/response`，全模块 handler 迁移完成 |
