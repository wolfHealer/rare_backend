
# 启动服务
go run cmd/server/main.go


# 
cmd/server/main.go —— 真正的启动入口（规范做法）
internal/config —— 配置集中管理
internal/router —— 路由统一注册
internal/module/* —— 按业务拆模块（auth / post）
internal/middleware —— 日志、鉴权、限流的入口位


handler（HTTP）

service（业务）
repo（数据库）
knowledge:知识库
post: 交流区
resource:资源


community：病友交流区


go get github.com/golang-jwt/jwt/v5





