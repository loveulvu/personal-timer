# Personal Study Timer

## Windows 开发环境启动

确保 MySQL 已经运行，并且本机已安装项目所需的 Go 和 Wails 开发环境。

可以在资源管理器中双击项目根目录下的 `start-dev.bat` 启动开发环境。

也可以在项目根目录的 PowerShell 中运行：

```powershell
.\start-dev.bat
```

启动脚本会：

1. 打开一个保持运行的 PowerShell 窗口并启动 Go 后端。
2. 等待 2 秒。
3. 打开另一个保持运行的 PowerShell 窗口并运行 `wails dev`。

脚本不会自动启动 MySQL。
