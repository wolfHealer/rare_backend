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

数据库表结构由 **RDS / 运维侧单独维护**，本仓库不包含建表 SQL 或 migration 脚本。部署前请确保 MySQL 中已有 `rare_backend` 库及业务所需表。

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
├── 启动顺序 + 配置外置（config/env） ✅
├── JWT 校验中间件 + 管理接口鉴权/角色 ✅
├── community：post_like 字段、DeletePost Scan、去掉 user_id 硬编码/body 伪造 ✅
└── 密钥/密码移出仓库 ✅

第二阶段（P1，2–4 周）
├── 引入 middleware（日志、recovery、auth）
├── 按模块拆分 handler → service → repo（先从 auth、community 试点）✅
├── 统一 API 响应契约（code/message/data）✅

第三阶段（P2） 
├── 消除 N+1（resource 列表 + 社区点赞态批量查询）✅
├── 列表去掉大字段、点赞计数改为增量 # todo
└── 索引与慢查询（LIKE、JOIN）✅ — `internal/pkg/search`

第四阶段（P3）
├── 去 fmt.Printf、统一命名与错误处理 — fmt.Printf 已清理，统一用 log.Printf
├── 抽公共 pagination/sqlbuilder/response # todo
└── 补测试与 CI # todo



变更说明
1. internal/module/auth/service/sms_service.go

新增 sendSMSPersonal：调用号码认证 SendSmsVerifyCode（dypnsapi.aliyuncs.com）
原 sendSMS 改名为 sendSMSCorporate 并整段注释，未删除
SendCode 现走个人资质接口；切回公司资质时取消注释即可
2. internal/config/config.go

新增配置项：SMS_PNVS_SIGN_NAME、SMS_PNVS_TEMPLATE_CODE
3. .env.example

补充个人资质配置说明；公司资质项改为注释示例
你需要在 .env 中配置
SMS_ACCESS_KEY_ID=你的AccessKeyId
SMS_ACCESS_KEY_SECRET=你的AccessKeySecret
# 号码认证控制台 → 短信认证 → 赠送签名/模板
SMS_PNVS_SIGN_NAME=速通互联验证码
SMS_PNVS_TEMPLATE_CODE=100001
签名与模板须为控制台配套赠送资源，模板变量名需与控制台一致（当前按 code、min 传参，有效期 5 分钟）。

开通入口：短信认证 - 赠送模板配置

RAM 权限需包含 dypns:SendSmsVerifyCode。

切回公司资质
在 SendCode 中注释 sendSMSPersonal，取消 sendSMSCorporate 注释，并配置 SMS_SIGN_NAME、SMS_REGISTER_TPL、SMS_LOGIN_TPL。





/api/community/posts/{id}/report:
    post:
      tags: [Community]
      summary: 举报帖子（合规）
      parameters:
        - $ref: '#/components/parameters/PathId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ReportPostRequest'
      responses:
        '200':
          description: 举报已提交
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponseNull'，现有CREATE TABLE `post` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '发帖用户ID',
  `disease_id` bigint DEFAULT NULL COMMENT '关联病种ID，可为空',
  `category_id` bigint DEFAULT NULL COMMENT '关联疾病分类ID，关联category表，可为空',
  `type` varchar(20) NOT NULL COMMENT '帖子类型：help/experience/emotion/info',
  `title` varchar(100) DEFAULT NULL COMMENT '标题，可为空（移动端发帖可不填）',
  `content` text NOT NULL COMMENT '正文内容',
  `images` json DEFAULT NULL COMMENT '图片URL数组',
  `view_count` int NOT NULL DEFAULT '0' COMMENT '浏览量',
  `like_count` int NOT NULL DEFAULT '0' COMMENT '点赞数',
  `comment_count` int NOT NULL DEFAULT '0' COMMENT '评论数',
  `favorite_count` int NOT NULL DEFAULT '0' COMMENT '收藏数',
  `is_top` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶：0否 1是',
  `is_recommend` tinyint NOT NULL DEFAULT '0' COMMENT '是否推荐：0否 1是',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态：0审核中 1正常 2驳回 3删除',
  `reject_reason` varchar(200) DEFAULT NULL COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_post_user_id` (`user_id`),
  KEY `idx_post_disease_id` (`disease_id`),
  KEY `idx_post_category_id` (`category_id`),
  KEY `idx_post_type` (`type`),
  KEY `idx_post_status` (`status`),
  KEY `idx_post_created_at` (`created_at`),
  KEY `idx_post_top_recommend` (`is_top`,`is_recommend`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='社区帖子表';CREATE TABLE `post_favorite` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL COMMENT '帖子ID',
  `user_id` bigint NOT NULL COMMENT '收藏用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_favorite_user` (`post_id`,`user_id`),
  KEY `idx_post_favorite_user_id` (`user_id`),
  KEY `idx_post_favorite_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='帖子收藏表';CREATE TABLE `post_like` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL COMMENT '帖子ID',
  `user_id` bigint NOT NULL COMMENT '点赞用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_user` (`post_id`,`user_id`),
  KEY `idx_post_like_user_id` (`user_id`),
  KEY `idx_post_like_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='帖子点赞表'，如果要实现举报，是否需要新增一张post_report表