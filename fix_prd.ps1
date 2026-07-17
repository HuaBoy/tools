$file = 'E:\tools20260623\src\views\tools\PrdGenerator.vue'
$content = [System.IO.File]::ReadAllText($file, [System.Text.Encoding]::UTF8)

# Fix 1: toArr - check Array.isArray BEFORE calling .trim()
$content = $content -replace "if \(!val \|\| val\.trim\(\) === '【待补充】'\) return \[\]", "if (Array.isArray(val)) return val; if (!val) return []; const _s = String(val); if (_s.trim() === '【待补充】') return []"

# Fix 2: dataRequirements - handle array case in downloadDocx
$content = $content -replace 'p\.dataRequirements \? p\.dataRequirements\.split', '(Array.isArray(p.dataRequirements) ? p.dataRequirements.join(''\n'') : p.dataRequirements).split'

[System.IO.File]::WriteAllText($file, $content, [System.Text.Encoding]::UTF8)
Write-Host "Done fixing PrdGenerator.vue"
