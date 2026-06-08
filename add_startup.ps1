$WshShell = New-Object -ComObject WScript.Shell
$StartupPath = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\icloud-code.lnk"
$Shortcut = $WshShell.CreateShortcut($StartupPath)
$Shortcut.TargetPath = "C:\Users\rahul\Documents\prog\go\icloudsolver\icloud-code.exe"
$Shortcut.WorkingDirectory = "C:\Users\rahul\Documents\prog\go\icloudsolver"
$Shortcut.Save()
Write-Host "Added to startup: $StartupPath"
