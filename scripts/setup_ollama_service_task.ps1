# 创建 Ollama 自愈任务：主进程直接是 ollama.exe serve，崩溃后自动重启
$ErrorActionPreference = 'Stop'
$svc = New-Object -ComObject Schedule.Service
$svc.Connect()
$root = $svc.GetFolder('\')
$taskName = 'OllamaServer'

# 清理旧任务
@('OllamaServer','OllamaWatchdog','OllamaService') | ForEach-Object {
    try { $root.DeleteTask($_, 0) } catch {}
}

$td = $svc.NewTask(0)
$td.RegistrationInfo.Description = 'Ollama 服务(ollama.exe serve)，崩溃后自动重启'
$td.RegistrationInfo.Author = '800228'

# 主体：当前交互用户
$p = $td.Principal
$p.UserId = '800228'
$p.LogonType = 3          # TASK_LOGON_INTERACTIVE_TOKEN
$p.RunLevel = 1           # 最高权限

# 设置：失败后每 1 分钟重启，最多 10 次；不限运行时长；已运行则不重复启动
$s = $td.Settings
$s.RestartCount = 10
$s.RestartInterval = 'PT1M'
$s.ExecutionTimeLimit = 'PT0S'
$s.MultipleInstances = 1  # TASK_INSTANCES_IGNORE_NEW：已运行则不重复启动
$s.DisallowStartIfOnBatteries = $false
$s.StopIfGoingOnBatteries = $false
$s.StartWhenAvailable = $true
# 关键：闲置时不要停止 ollama
$s.IdleSettings.StopOnIdleEnd = $false
$s.IdleSettings.RestartOnIdle = $false

# 动作：直接运行 ollama.exe serve（主进程就是它，任务结束才结束）
$a = $td.Actions.Create(0)   # TASK_ACTION_EXEC
$a.Path = 'C:\Users\800228\AppData\Local\Programs\Ollama\ollama.exe'
$a.Arguments = 'serve'
$a.WorkingDirectory = 'C:\Users\800228\AppData\Local\Programs\Ollama'

# 触发器：登录时启动（崩溃由 RestartOnFailure 兜底，避免双触发器抢端口）
$t1 = $td.Triggers.Create(9)  # TASK_TRIGGER_LOGON

$root.RegisterTaskDefinition($taskName, $td, 6, $null, $null, 3)  # CREATE_OR_UPDATE + logon
Write-Host "Task '$taskName' created."
