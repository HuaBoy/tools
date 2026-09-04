# Ollama 自愈看门狗：检测 11434 端口，若未监听则拉起 ollama serve
$ErrorActionPreference = 'SilentlyContinue'
$port = 11434
$ollamaExe = 'C:\Users\800228\AppData\Local\Programs\Ollama\ollama.exe'
$log = 'C:\temp\ollama_watchdog.log'

$listening = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
if ($listening) {
    # 端口正常，无需操作
    exit 0
}

# 端口未监听 -> 尝试拉起
try {
    Start-Process -FilePath $ollamaExe -ArgumentList 'serve' -WindowStyle Hidden -PassThru | Out-Null
    $msg = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') RESTART ollama serve (port $port was down)"
    Add-Content -Path $log -Value $msg
} catch {
    $msg = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') FAILED to start ollama: $_"
    Add-Content -Path $log -Value $msg
}
