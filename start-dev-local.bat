@echo off
chcp 65001 >nul
echo ============================================
echo   启动前端（本地后端代理模式）
echo   API 代理: http://localhost:8080
echo   访问地址: http://localhost:5173
echo ============================================
set VITE_PROXY_TARGET=http://localhost:8080
cd /d E:\tools20260623
npm run dev
