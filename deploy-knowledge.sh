#!/bin/bash
# =====================================================
# 内部知识库 + RAG + AI 对话 一键部署脚本
# 用法: bash deploy-knowledge.sh
# =====================================================
set -e

echo "=============================================="
echo "  内部知识库系统部署脚本"
echo "=============================================="

# 1. 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "[1/6] 未检测到 Docker，开始安装..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
else
    echo "[1/6] Docker 已安装: $(docker --version)"
fi

# 2. 检查 Compose 插件
if ! docker compose version &> /dev/null; then
    echo "[2/6] 安装 Docker Compose 插件..."
    apt update && apt install -y docker-compose-plugin
else
    echo "[2/6] Docker Compose 已安装"
fi

# 3. 启动 ollama 并拉取模型
echo "[3/6] 启动 Ollama 服务..."
docker compose up -d ollama

# 等待 ollama 就绪
echo "等待 Ollama 就绪..."
for i in $(seq 1 30); do
    if docker exec ai-kb-ollama ollama list &> /dev/null; then
        break
    fi
    sleep 2
done

echo "[4/6] 拉取模型（deepseek-r1:7b 约 4.7GB，请耐心等待）..."
docker exec ai-kb-ollama ollama pull deepseek-r1:7b || echo "警告: deepseek-r1:7b 拉取失败，请手动执行 docker exec ai-kb-ollama ollama pull deepseek-r1:7b"
docker exec ai-kb-ollama ollama pull bge-small-zh-v1.5 || echo "警告: bge-small-zh-v1.5 拉取失败，请手动执行"

echo "[5/6] 启动全部服务..."
docker compose up -d --build

# 4. 验证
echo "[6/6] 验证服务..."
sleep 5
docker compose ps

echo ""
echo "=============================================="
echo "  部署完成！"
echo "  前端入口:  http://$(hostname -I 2>/dev/null | awk '{print $1}' | head -n1)/"
echo "  健康检查:  http://localhost:8080/health"
echo ""
echo "  模型列表:"
docker exec ai-kb-ollama ollama list 2>/dev/null || echo "  (ollama 未就绪，稍后执行 docker exec ai-kb-ollama ollama list)"
echo "=============================================="
