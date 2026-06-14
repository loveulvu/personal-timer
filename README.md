# Personal Study Timer

## Windows 开发环境启动

确保 MySQL 已经运行，并且本机已安装项目所需的 Go 和 Wails 开发环境。

推荐在资源管理器中双击项目根目录下的 `start-desktop-dev.bat` 启动开发环境。

也可以从任意路径在 PowerShell 中运行：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File E:\Projects\personal-study-timer\scripts\start-desktop-dev.ps1
```

启动脚本会：

1. 检查项目目录、`.env`、Go、npm 和 Wails CLI。
2. 仅在缺少 `frontend/node_modules` 时执行 `npm install`。
3. 运行 TypeScript 检查和最多 60 秒的前端构建检查。
4. 打开一个保持运行的 PowerShell 窗口启动 Go 后端。
5. 在当前窗口运行 `wails dev`。

脚本不会自动启动 MySQL。

原有的 `start-dev.bat` 会转发到推荐启动脚本。
