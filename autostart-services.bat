@echo off
chcp 65001 >nul
echo ==========================================
echo   知识库服务 - 开机自启动（静默模式）
echo   %date% %time%
echo ==========================================

set ROOT=E:\tools20260623
set LOG_DIR=%ROOT%\logs

:: 创建日志目录
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

:: ===== 1. 等待系统就绪 =====
echo [1/5] 等待系统就绪 (10秒)...
timeout /t 10 /nobreak >nul

:: ===== 2. 确保 PostgreSQL 在运行 =====
echo [2/5] 检查 PostgreSQL...
sc query postgresql-x64-16 | find "RUNNING" >nul
if errorlevel 1 (
    echo     PostgreSQL 未运行，正在启动...
    net start postgresql-x64-16 >nul 2>&1
    timeout /t 5 /nobreak >nul
)
echo     PostgreSQL 已就绪

:: ===== 3. 启动 Ollama =====
echo [3/5] 启动 Ollama...
tasklist /FI "IMAGENAME eq ollama.exe" | find /I "ollama" >nul
if errorlevel 1 (
    echo     Ollama 未运行，正在启动...
    start "Ollama" /min ollama serve
    timeout /t 5 /nobreak >nul
) else (
    echo     Ollama 已在运行
)

:: ===== 4. 启动 Python RAG 服务 =====
echo [4/5] 启动 Python LlamaIndex RAG 服务 (端口 8001)...
tasklist /FI "WINDOWTITLE eq Python-RAG" | find /I "cmd" >nul
if errorlevel 1 (
    start "Python-RAG" /min cmd /c "cd /d %ROOT%\python-rag && python run.py >> %LOG_DIR%\python-rag.log 2>&1"
    timeout /t 5 /nobreak >nul
    echo     Python RAG 服务已启动
) else (
    echo     Python RAG 已在运行
)

:: ===== 5. 启动 Go 后端 =====
echo [5/5] 启动 Go 后端 (端口 8080)...
tasklist /FI "WINDOWTITLE eq Go-Backend" | find /I "cmd" >nul
if errorlevel 1 (
    start "Go-Backend" /min cmd /c "cd /d %ROOT%\backend && server.exe >> %LOG_DIR%\go-backend.log 2>&1"
    timeout /t 3 /nobreak >nul
    echo     Go 后端已启动
) else (
    echo     Go 后端已在运行
)

:: ===== 验证服务状态 =====
echo.
echo ==========================================
echo   服务状态验证:
echo ==========================================

:: 验证 PostgreSQL
sc query postgresql-x64-16 | find "RUNNING" >nul
if errorlevel 1 (echo   [X] PostgreSQL - 未运行) else (echo   [OK] PostgreSQL - 运行中)

:: 验证 Ollama
tasklist /FI "IMAGENAME eq ollama.exe" | find /I "ollama" >nul
if errorlevel 1 (echo   [X] Ollama - 未运行) else (echo   [OK] Ollama - 运行中)

:: 验证 Python RAG (检查 8001 端口)
netstat -an | find ":8001" | find "LISTENING" >nul
if errorlevel 1 (echo   [X] Python RAG - 未运行 (端口 8001) ) else (echo   [OK] Python RAG - 运行中 (端口 8001))

:: 验证 Go 后端 (检查 8080 端口)
netstat -an | find ":8080" | find "LISTENING" >nul
if errorlevel 1 (echo   [X] Go 后端 - 未运行 (端口 8080) ) else (echo   [OK] Go 后端 - 运行中 (端口 8080))

echo ==========================================
echo   所有服务启动完成！
echo   前端访问: http://localhost:3000
echo ==========================================
