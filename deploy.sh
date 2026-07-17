#!/bin/bash

set -e

echo "=========================================="
echo "AI起爆一体化工具集 - 部署脚本"
echo "=========================================="

if [ -z "$1" ]; then
    echo "用法: $0 <服务器IP>"
    exit 1
fi

SERVER_IP="$1"

echo ""
echo "步骤 1: 创建项目目录并清理旧文件"
echo "------------------------------------------"
ssh root@$SERVER_IP "mkdir -p /www/wwwroot/tester-platform && rm -rf /www/wwwroot/tester-platform/*"

echo ""
echo "步骤 2: 上传代码（排除 node_modules/.git/dist）"
echo "------------------------------------------"
rsync -avz --progress \
  --exclude='node_modules' \
  --exclude='.git' \
  --exclude='dist' \
  --exclude='.DS_Store' \
  --exclude='*.log' \
  ./ root@$SERVER_IP:/www/wwwroot/tester-platform/

echo ""
echo "步骤 3: 更新环境变量"
echo "------------------------------------------"
ssh root@$SERVER_IP "sed -i 's/您的服务器IP/$SERVER_IP/g' /www/wwwroot/tester-platform/backend/.env"

echo ""
echo "步骤 4: 启动 Docker 容器"
echo "------------------------------------------"
ssh root@$SERVER_IP "cd /www/wwwroot/tester-platform && docker compose up -d --build"

echo ""
echo "步骤 5: 等待服务启动..."
sleep 15

echo ""
echo "步骤 6: 验证服务"
echo "------------------------------------------"
echo "前端服务:"
curl -s http://$SERVER_IP/health || echo "前端服务未启动"

echo ""
echo "后端服务:"
curl -s http://$SERVER_IP:8080/health || echo "后端服务未启动"

echo ""
echo "=========================================="
echo "部署完成!"
echo "访问地址: http://$SERVER_IP"
echo "=========================================="