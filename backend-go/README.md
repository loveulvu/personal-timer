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
- weekly stats
- manual daily summary generation
- manual weekly summary generation
- generated_summaries persistence

`daily_tasks` 的 PUT 接口接受 `planned`、`running`、`paused`、`completed`、
`cancelled` 五种合法状态。为避免绕过 timer 状态机，手动状态切换仅允许
`planned` 与 `cancelled` 互转；其他状态变化应使用 timer 接口。

## 启动方式

1. 在项目根目录配置 `.env` 中的 MySQL 连接信息。
2. 如果需要生成总结，配置 LLM 环境变量：

   ```text
   LLM_API_KEY=your_api_key_here
   LLM_BASE_URL=https://api.openai.com/v1
   LLM_MODEL=gpt-4o-mini
   ```

   `LLM_BASE_URL` 使用兼容 `/chat/completions` 的 HTTP API 地址，不要把真实
   API key 写入代码或文档。

3. 进入后端目录并启动：

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
- `GET /api/stats/weekly?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`
- `POST /api/summaries/daily/generate`
- `POST /api/summaries/weekly/generate`
- `GET /api/summaries?type=daily|weekly`
- `GET /api/summaries/:id`
- `DELETE /api/summaries/:id`

weekly stats 不新建 weekly 表，实时从 `daily_tasks`、`time_sessions`、
`projects` 聚合。当前不强制日期范围必须正好 7 天，但推荐传一周范围。

## Backend smoke test

当前项目仍然只做后端，最终计划接桌面端应用；暂不做 Web 前端、Redis、MQ、
登录、多用户、月统计或分类系统。

以下示例假设服务运行在 `http://localhost:8085`，并使用 `jq` 读取返回 id：

```bash
BASE_URL=http://localhost:8085
DATE=2026-06-08
WEEK_START=2026-06-08
WEEK_END=2026-06-14

curl "$BASE_URL/api/health"

PROJECT_ID=$(curl -s -X POST "$BASE_URL/api/projects" \
  -H "Content-Type: application/json" \
  -d '{"name":"Study","description":"Backend smoke test","is_fixed":false}' \
  | jq -r '.id')

TASK_ID=$(curl -s -X POST "$BASE_URL/api/daily-tasks" \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":$PROJECT_ID,\"task_date\":\"$DATE\",\"title\":\"Read docs\",\"estimated_minutes\":25}" \
  | jq -r '.id')

curl -X POST "$BASE_URL/api/daily-tasks/$TASK_ID/start"
curl -X POST "$BASE_URL/api/daily-tasks/$TASK_ID/pause"
curl -X POST "$BASE_URL/api/daily-tasks/$TASK_ID/resume"
curl -X POST "$BASE_URL/api/daily-tasks/$TASK_ID/finish"

curl "$BASE_URL/api/stats/daily?date=$DATE"
curl "$BASE_URL/api/stats/weekly?start_date=$WEEK_START&end_date=$WEEK_END"

DAILY_SUMMARY_ID=$(curl -s -X POST "$BASE_URL/api/summaries/daily/generate" \
  -H "Content-Type: application/json" \
  -d "{\"date\":\"$DATE\"}" \
  | jq -r '.data.summary_id')

WEEKLY_SUMMARY_ID=$(curl -s -X POST "$BASE_URL/api/summaries/weekly/generate" \
  -H "Content-Type: application/json" \
  -d "{\"start_date\":\"$WEEK_START\",\"end_date\":\"$WEEK_END\"}" \
  | jq -r '.data.summary_id')

curl "$BASE_URL/api/summaries?type=daily"
curl "$BASE_URL/api/summaries/$DAILY_SUMMARY_ID"
curl -X DELETE "$BASE_URL/api/summaries/$DAILY_SUMMARY_ID"
curl -X DELETE "$BASE_URL/api/summaries/$WEEKLY_SUMMARY_ID"
```

当前 README 是阶段性简略版本，后续桌面端完成后再重构完整文档。
