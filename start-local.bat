@echo off
chcp 65001 >nul
set VITE_PROXY_TARGET=http://localhost:8080
cd /d E:\tools20260623
node_modules\.bin\vite.cmd --host 0.0.0.0 --port 5173
