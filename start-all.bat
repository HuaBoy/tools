@echo off
chcp 65001 >nul
echo ==========================================
echo   AI起爆一体化工具集 - 全部服务启动脚本
echo ==========================================
echo.

set ROOT=%~dp0

:: 0. 确保 Ollama 在运行
echo [0/4] 检查 Ollama...
tasklist /FI "IMAGENAME eq ollama.exe" | find /I "ollama" >nul
if errorlevel 1 (
    echo     Ollama 未运行，正在启动...
    start "Ollama" /min ollama serve
    timeout /t 3 /nobreak >nul
)
echo     Ollama 已就绪 (端口 11434)

:: 1. 启动 Python RAG 服务 (端口 8001)
echo [1/4] 启动 Python LlamaIndex RAG 服务 (端口 8001)...
start "Python-RAG" /min cmd /c "cd /d %ROOT%python-rag && python run.py"
timeout /t 5 /nobreak >nul

:: 2. 启动 Go 后端 (端口 8080)
echo [2/4] 启动 Go 后端 (端口 8080)...
start "Go-Backend" /min cmd /c "cd /d %ROOT%backend && server.exe"
timeout /t 2 /nobreak >nul

:: 3. 启动前端 (端口 3000)
echo [3/4] 启动 Vue 前端 (端口 3000)...
start "Vue-Frontend" cmd /c "cd /d %ROOT% && npm run dev"

echo.
echo ==========================================
echo   全部服务已启动:
echo   - Ollama:        http://localhost:11434 (AI模型)
echo   - Python RAG:    http://localhost:8001  (LlamaIndex向量检索)
echo   - Go 后端:       http://localhost:8080  (API服务)
echo   - Vue 前端:      http://localhost:3000  (浏览器访问)
echo ==========================================
echo.
echo   浏览器打开 http://localhost:3000 即可使用
echo.
pause
