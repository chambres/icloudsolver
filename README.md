# icloudsolver

Auto-types the iCloud Passwords 2FA code on Windows so you don't have to.

When you log into something with iCloud Passwords, the Windows extension pops up
a little dialog with a 6-digit verification code that you then have to read and
type in by hand. This watches for that dialog and does it for you.

## How it works

- Polls for the `iCloudPasswordsExtensionHelper.exe` process and watches its
  windows (Win32 `EnumWindows` / `EnumChildWindows`).
- When a code dialog (`#32770`) goes from hidden to visible, it reads the child
  text controls and regexes out the `### ###` code.
- Types the code with simulated keystrokes via `keybd_event`.

## Run it

```bash
go run main.go
```

Or build once and let it run in the background:

```bash
go build -o icloud-code.exe
```

`add_startup.ps1` drops a shortcut in your Startup folder so it launches on
login (edit the path inside it to wherever you put the exe).

Windows only — it's all Win32 API.
