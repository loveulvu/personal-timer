# Personal Study Timer

Personal Study Timer 是个人学习与项目执行计时工具的后端项目。当前阶段只完成
后端 MVP，最终计划接入桌面端应用，不是 Web 前端。

## 技术栈

- Go
- Gin
- MySQL
- `database/sql`

## 当前核心功能

- projects CRUD
- daily_tasks CRUD
- timer start / pause / resume / finish
- daily stats

`daily_tasks` 的 PUT 接口接受 `planned`、`running`、`paused`、`completed`、
`cancelled` 五种合法状态。为避免绕过 timer 状态机，手动状态切换仅允许
`planned` 与 `cancelled` 互转；其他状态变化应使用 timer 接口。

## 启动方式

1. 在项目根目录配置 `.env` 中的 MySQL 连接信息。
2. 进入后端目录并启动：

   ```bash
   cd backend-go
   go run ./cmd/server
   ```

服务默认监听 `http://localhost:8085`。

## 主要接口

- `GET /api/health`
- `GET /api/health/db`
- `POST /api/projects`
- `GET /api/projects`
- `GET /api/projects/:id`
- `PUT /api/projects/:id`
- `DELETE /api/projects/:id`
- `POST /api/daily-tasks`
- `GET /api/daily-tasks?date=YYYY-MM-DD`
- `GET /api/daily-tasks/:id`
- `PUT /api/daily-tasks/:id`
- `DELETE /api/daily-tasks/:id`
- `POST /api/daily-tasks/:id/start`
- `POST /api/daily-tasks/:id/pause`
- `POST /api/daily-tasks/:id/resume`
- `POST /api/daily-tasks/:id/finish`
- `GET /api/stats/daily?date=YYYY-MM-DD`

当前 README 是阶段性简略版本，后续桌面端完成后再重构完整文档。
