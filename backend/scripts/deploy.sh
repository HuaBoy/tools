#!/bin/bash
# ========================================
# 阿里云 CentOS 7 部署脚本
# ========================================

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
APP_NAME="tester-platform"
APP_DIR="/www/wwwroot/${APP_NAME}"
BACKEND_DIR="${APP_DIR}/backend"
BINARY_NAME="tester-platform"
SERVICE_NAME="${APP_NAME}.service"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ${APP_NAME} 部署脚本 (CentOS 7)${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 root 用户运行此脚本${NC}"
    exit 1
fi

# 1. 关闭 SELinux
echo -e "${YELLOW}[1/10] 配置 SELinux...${NC}"
setenforce 0 2>/dev/null || true
sed -i 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config 2>/dev/null || true
echo -e "${GREEN}SELinux 已关闭${NC}"

# 2. 检查系统
echo -e "${YELLOW}[2/10] 检查系统...${NC}"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo -e "系统: ${GREEN}${PRETTY_NAME}${NC}"
    if [[ "$ID" != "centos" ]] && [[ "$ID" != "rhel" ]] && [[ "$ID" != "almalinux" ]] && [[ "$ID" != "rocky" ]]; then
        echo -e "${YELLOW}警告: 此脚本专为 CentOS/RHEL 设计，当前系统为 $ID${NC}"
    fi
fi

# 3. 安装基础工具
echo -e "${YELLOW}[3/10] 安装基础工具...${NC}"
yum install -y epel-release yum-utils wget curl git vim firewalld 2>/dev/null || true
yum install -y supervisor 2>/dev/null || true

# 4. 启动 firewalld
echo -e "${YELLOW}[4/10] 配置防火墙...${NC}"
systemctl enable firewalld 2>/dev/null || true
systemctl start firewalld 2>/dev/null || true
firewall-cmd --permanent --add-service=ssh 2>/dev/null || true
firewall-cmd --permanent --add-service=http 2>/dev/null || true
firewall-cmd --permanent --add-service=https 2>/dev/null || true
firewall-cmd --reload 2>/dev/null || true
echo -e "${GREEN}防火墙配置完成${NC}"

# 5. 安装 Go
echo -e "${YELLOW}[5/10] 检查 Go 环境...${NC}"
if ! command -v go &> /dev/null; then
    echo "正在安装 Go 1.21.5..."
    cd /tmp
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo 'export GOPROXY=https://goproxy.cn,direct' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    export GOPROXY=https://goproxy.cn,direct
    echo -e "${GREEN}Go 安装完成: $(/usr/local/go/bin/go version)${NC}"
else
    echo -e "${GREEN}Go 已安装: $(go version)${NC}"
fi

# 6. 创建 www 用户
echo -e "${YELLOW}[6/10] 创建 www 用户...${NC}"
if ! id "www" &>/dev/null; then
    useradd -r -s /bin/false www
    echo -e "${GREEN}www 用户创建完成${NC}"
fi

# 7. 创建项目目录
echo -e "${YELLOW}[7/10] 创建项目目录...${NC}"
mkdir -p ${BACKEND_DIR}/{logs,bin,configs}
chown -R www:www ${APP_DIR}

# 8. 配置环境变量
echo -e "${YELLOW}[8/10] 检查环境变量...${NC}"
if [ ! -f "${BACKEND_DIR}/.env" ]; then
    cp ${BACKEND_DIR}/.env.example ${BACKEND_DIR}/.env
    echo -e "${RED}⚠️  请编辑 ${BACKEND_DIR}/.env 填入真实的 Supabase 配置${NC}"
    echo -e "${YELLOW}按任意键打开编辑器，或 Ctrl+C 退出后手动编辑...${NC}"
    read -n 1
    vim ${BACKEND_DIR}/.env
fi

# 9. 编译
echo -e "${YELLOW}[9/10] 编译 Go 程序...${NC}"
cd ${BACKEND_DIR}
export PATH=$PATH:/usr/local/go/bin
export GOPROXY=https://goproxy.cn,direct
go mod tidy
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/${BINARY_NAME} ./cmd/server
chmod +x bin/${BINARY_NAME}
chown -R www:www ${BACKEND_DIR}
echo -e "${GREEN}编译完成: bin/${BINARY_NAME} ($(du -h bin/${BINARY_NAME} | cut -f1))${NC}"

# 10. 配置 systemd 服务
echo -e "${YELLOW}[10/10] 配置 systemd 服务...${NC}"
cp ${BACKEND_DIR}/scripts/${SERVICE_NAME} /etc/systemd/system/${SERVICE_NAME}

# 修复 service 文件中的路径
sed -i "s|/www/wwwroot/tester-platform/backend|${BACKEND_DIR}|g" /etc/systemd/system/${SERVICE_NAME}

systemctl daemon-reload
systemctl enable ${SERVICE_NAME}
systemctl restart ${SERVICE_NAME}
sleep 2
systemctl status ${SERVICE_NAME} --no-pager

# 安装 Nginx（如果未安装）
if ! command -v nginx &> /dev/null; then
    echo -e "${YELLOW}安装 Nginx...${NC}"
    rpm -Uvh http://nginx.org/packages/centos/7/noarch/RPMS/nginx-release-centos-7-0.el7.ngx.noarch.rpm 2>/dev/null || true
    yum install -y nginx
    systemctl enable nginx
    systemctl start nginx
    echo -e "${GREEN}Nginx 安装完成${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "服务状态: ${YELLOW}systemctl status ${SERVICE_NAME}${NC}"
echo -e "查看日志: ${YELLOW}journalctl -u ${SERVICE_NAME} -f${NC}"
echo -e "重启服务: ${YELLOW}systemctl restart ${SERVICE_NAME}${NC}"
echo -e "健康检查: ${YELLOW}curl http://127.0.0.1:8080/health${NC}"
echo ""
echo -e "${YELLOW}后续步骤：${NC}"
echo -e "1. 在阿里云控制台安全组开放 80/443 端口"
echo -e "2. 配置 Nginx 反向代理："
echo -e "   ${YELLOW}cp ${BACKEND_DIR}/scripts/nginx.conf.example /etc/nginx/conf.d/${APP_NAME}.conf${NC}"
echo -e "   ${YELLOW}vim /etc/nginx/conf.d/${APP_NAME}.conf${NC}  (修改 server_name)"
echo -e "   ${YELLOW}nginx -t && systemctl reload nginx${NC}"
echo -e "3. 在 Supabase Dashboard 中执行 scripts/init-supabase.sql"
echo -e "4. 创建管理员用户: ${YELLOW}SELECT public.set_admin('your-email@example.com');${NC}"
echo ""
