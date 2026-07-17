#!/bin/bash
# 阿里云服务器初始化脚本（首次部署使用）
# 在全新的阿里云 ECS 上运行

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  阿里云服务器初始化${NC}"
echo -e "${GREEN}========================================${NC}"

# 1. 更新系统
echo -e "${YELLOW}[1/7] 更新系统...${NC}"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    if [[ "$ID" == "centos" ]] || [[ "$ID" == "rhel" ]] || [[ "$ID" == "almalinux" ]]; then
        yum update -y
        yum install -y epel-release yum-utils
    else
        apt update -y && apt upgrade -y
    fi
fi

# 2. 安装基础工具
echo -e "${YELLOW}[2/7] 安装基础工具...${NC}"
if command -v apt &> /dev/null; then
    apt install -y curl wget git vim ufw supervisor
elif command -v yum &> /dev/null; then
    yum install -y curl wget git vim policycoreutils-python-utils
fi

# 3. 创建 www 用户
echo -e "${YELLOW}[3/7] 创建 www 用户...${NC}"
if ! id "www" &>/dev/null; then
    useradd -r -s /bin/false www
    echo -e "${GREEN}www 用户创建完成${NC}"
fi

# 4. 创建项目目录
echo -e "${YELLOW}[4/7] 创建项目目录...${NC}"
mkdir -p /www/wwwroot
chown -R www:www /www

# 5. 安装 Nginx
echo -e "${YELLOW}[5/7] 安装 Nginx...${NC}"
if ! command -v nginx &> /dev/null; then
    if command -v apt &> /dev/null; then
        apt install -y nginx
    elif command -v yum &> /dev/null; then
        yum install -y nginx
    fi
    systemctl enable nginx
    systemctl start nginx
    echo -e "${GREEN}Nginx 安装完成${NC}"
fi

# 6. 安装 Certbot（用于 HTTPS）
echo -e "${YELLOW}[6/7] 安装 Certbot...${NC}"
if ! command -v certbot &> /dev/null; then
    if command -v apt &> /dev/null; then
        apt install -y certbot python3-certbot-nginx
    elif command -v yum &> /dev/null; then
        yum install -y certbot python3-certbot-nginx
    fi
fi

# 7. 配置防火墙
echo -e "${YELLOW}[7/7] 配置防火墙...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow 22/tcp
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw --force enable
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-service=ssh
    firewall-cmd --permanent --add-service=http
    firewall-cmd --permanent --add-service=https
    firewall-cmd --reload
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  初始化完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}重要提示：${NC}"
echo -e "1. ${RED}请在阿里云控制台安全组中开放 80/443 端口${NC}"
echo -e "2. ${RED}阿里云安全组优先级高于系统防火墙${NC}"
echo -e "3. 将代码上传到 /www/wwwroot/tester-platform/"
echo -e "4. 运行 deploy.sh 部署应用"
echo ""
echo -e "${YELLOW}阿里云安全组配置：${NC}"
echo -e "   入方向 → 添加安全组规则"
echo -e "   - 端口范围: 80/80, 协议: TCP"
echo -e "   - 端口范围: 443/443, 协议: TCP"
echo -e "   - 端口范围: 22/22, 协议: TCP (SSH)"
echo -e "   - 授权对象: 0.0.0.0/0"
echo ""
