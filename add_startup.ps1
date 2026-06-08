$WshShell = New-Object -ComObject WScript.Shell
$StartupPath = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\icloud-code.lnk"
$Shortcut = $WshShell.CreateShortcut($StartupPath)
$Shortcut.TargetPath = Join-Path $PSScriptRoot "icloud-code.exe"
$Shortcut.WorkingDirectory = $PSScriptRoot
$Shortcut.Save()
Write-Host "Added to startup: $StartupPath"
