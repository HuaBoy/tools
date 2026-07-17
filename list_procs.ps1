$procs = Get-CimInstance Win32_Process -Filter "Name = 'node.exe'" -ErrorAction SilentlyContinue
$conns = Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue
foreach ($p in $procs) {
  $pidx = $p.ProcessId
  $ports = ($conns | Where-Object { $_.OwningProcess -eq $pidx } | Select-Object -ExpandProperty LocalPort) -join ','
  $cmd = $p.CommandLine
  if ($cmd -match 'vite|npm|node') {
    Write-Host "PID=$pidx PORTS=$ports"
    Write-Host "   CMD=$cmd"
  }
}
