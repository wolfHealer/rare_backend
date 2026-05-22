# 数据库 Migration

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 管理 schema 版本。

## 文件说明

| 文件 | 用途 |
|------|------|
| `schema_reference.sql` | mysqldump 结构快照，**只读参考**，勿在有数据的库上执行 |
| `000001_baseline.up.sql` | 初始建表（由 reference 自动生成） |
| `000001_baseline.down.sql` | 回滚 baseline（DROP 全部表，仅 dev/test） |
| `000002_*.up/down.sql` | 列表/JOIN 复合索引 + FULLTEXT（ngram）关键词搜索 |

重新生成 baseline：

```bash
make migrate-gen-baseline
```

## 首次使用

```bash
cp .env.example .env          # 配置 MYSQL_DSN
make migrate-install          # 安装 migrate CLI
make migrate-up               # 建表
```

## 常用命令

```bash
make migrate-up               # 执行未应用的 migration
make migrate-down             # 回滚一步
make migrate-version          # 查看当前版本
make migrate-create NAME=add_foo   # 新建 000002_add_foo.up/down.sql
make db-reset                 # 清空并重建（开发环境）
```

## 新增变更流程

1. `make migrate-create NAME=describe_change`
2. 编辑生成的 `.up.sql` / `.down.sql`
3. 本地 `make migrate-up` 验证
4. 提交 migration 文件，部署环境执行 `make migrate-up`

## 注意

- baseline **不含 seed 数据**，业务数据需另行导入。
- `MYSQL_DSN` 密码含特殊字符时，可能需 URL 编码后再用于 migrate。
- 生产环境禁止 `make db-reset` 与 `000001_baseline.down.sql`。
