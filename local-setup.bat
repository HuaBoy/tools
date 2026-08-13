@echo off
chcp 65001 >nul
title 本地知识库环境搭建
echo ==================================================
echo   本地知识库 + AI 环境搭建脚本
echo   （安装 Ollama / PostgreSQL，拉取模型）
echo ==================================================
echo.

:: ================= 1. 安装 Ollama =================
echo [1/5] 检查 Ollama...
where ollama >nul 2>&1
if %errorlevel% equ 0 (
    echo       Ollama 已安装:
    ollama --version
) else (
    echo       未安装，开始安装（下载约 700MB）...
    echo       如果弹出 UAC 窗口，请点击"是"
    winget install Ollama.Ollama --accept-source-agreements --accept-package-agreements
    if %errorlevel% neq 0 (
        echo [错误] Ollama 安装失败！请手动下载安装:
        echo       https://ollama.com/download/windows
        pause
        exit /b 1
    )
)

:: ================= 2. 拉取模型 =================
echo.
echo [2/5] 拉取 AI 模型（deepseek-r1:7b 约 4.7GB，根据网速可能需要 10-30 分钟）...
echo       期间请保持电脑不锁屏、不断网。
echo.
ollama pull deepseek-r1:7b
ollama pull bge-small-zh-v1.5
echo.
echo       模型拉取完成！已安装模型：
ollama list

:: ================= 3. 安装 PostgreSQL 16 =================
echo.
echo [3/5] 检查 PostgreSQL...
where psql >nul 2>&1
if %errorlevel% equ 0 (
    echo       PostgreSQL 已安装:
    psql --version
) else (
    echo       未安装，开始安装（下载约 300MB）...
    echo       【重要】安装过程中会要求设置 postgres 超级用户密码
    echo       请记住你设置的密码（后面要用）
    echo       如果弹出 UAC 窗口，请点击"是"
    winget install PostgreSQL.PostgreSQL.16 --accept-source-agreements --accept-package-agreements
    if %errorlevel% neq 0 (
        echo [错误] PostgreSQL 安装失败！
        echo       请手动安装: https://www.enterprisedb.com/downloads/postgres-postgresql-downloads
        echo       安装时选择 PostgreSQL 16 版本，记好密码
        pause
        exit /b 1
    )
)

:: ================= 4. 初始化数据库 =================
echo.
echo [4/5] 初始化数据库...
set PGPATH=C:\Program Files\PostgreSQL\16\bin

"%PGPATH%\createdb.exe" -U postgres -h localhost knowledge 2>nul
if %errorlevel% neq 0 (
    echo       数据库可能已存在，继续...
)

"%PGPATH%\psql.exe" -U postgres -h localhost -d knowledge -c "CREATE EXTENSION IF NOT EXISTS vector;" 2>nul
"%PGPATH%\psql.exe" -U postgres -h localhost -d knowledge -c "CREATE USER kbuser WITH PASSWORD 'kbpass_2026';" 2>nul
"%PGPATH%\psql.exe" -U postgres -h localhost -d knowledge -c "GRANT ALL PRIVILEGES ON DATABASE knowledge TO kbuser;" 2>nul
"%PGPATH%\psql.exe" -U postgres -h localhost -d knowledge -c "GRANT ALL ON SCHEMA public TO kbuser;" 2>nul

echo       数据库初始化完成。

:: ================= 5. 完成提示 =================
echo.
echo [5/5] 环境搭建完成！
echo ==================================================
echo   接下来请按以下顺序启动：
echo.
echo   1. 修改 backend\.env 中的数据库连接（如未配置）:
echo      DB_DSN=postgres://kbuser:kbpass_2026@localhost:5432/knowledge?sslmode=disable
echo.
echo   2. 启动后端（新终端窗口）:
echo      cd E:\tools20260623\backend
echo      go run ./cmd/server
echo.
echo   3. 启动前端（新终端窗口）:
echo      cd E:\tools20260623
echo      npm run dev
echo.
echo   4. 浏览器打开 http://localhost:5173
echo      登录账号: admin  密码: 800228
echo ==================================================
echo.
pause
