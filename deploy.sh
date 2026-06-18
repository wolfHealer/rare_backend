#!/bin/bash
# 本地一键部署到 ECS（仅覆盖 rare_server，不触碰服务器 .env）
set -e

ECS_IP="${ECS_IP:-112.74.103.140}"
SSH_USER="${SSH_USER:-root}"
REMOTE_DIR="${REMOTE_DIR:-/opt/rare_backend}"

# 可选：在本机创建 deploy.local.env（已 gitignore），例如：
#   SSH_USER=root
#   ECS_IP=112.74.103.140
# 勿将密码写入任何会提交的文件；推荐配置 SSH 公钥免密登录。
if [[ -f deploy.local.env ]]; then
  # shellcheck disable=SC1091
  source deploy.local.env
fi

SSH_TARGET="${SSH_USER}@${ECS_IP}"
SCP_OPTS=()
SSH_OPTS=()

cd "$(dirname "$0")"

echo "==> 目标: ${SSH_TARGET}:${REMOTE_DIR}"

echo "==> 编译 Linux 二进制..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o rare_server cmd/server/main.go

echo "==> 上传 rare_server（临时文件，避免覆盖运行中二进制失败）..."
scp "${SCP_OPTS[@]}" rare_server "${SSH_TARGET}:${REMOTE_DIR}/rare_server.new"

echo "==> 替换并重启服务..."
ssh "${SSH_OPTS[@]}" "${SSH_TARGET}" "
  set -e
  systemctl stop rare_backend || true
  mv -f ${REMOTE_DIR}/rare_server.new ${REMOTE_DIR}/rare_server
  chmod +x ${REMOTE_DIR}/rare_server
  systemctl start rare_backend
  sleep 1
  systemctl is-active rare_backend
"

echo "==> 本机探活..."
ssh "${SSH_OPTS[@]}" "${SSH_TARGET}" "curl -s http://127.0.0.1:8080/health"
echo ""
echo "部署完成: https://api.rarelink.com.cn/health"
