# Personal Study Timer Backend

Personal Study Timer is a small study-planning and time-tracking API. It manages
projects and daily tasks, records multiple study sessions for each task, and
provides daily statistics.

## Tech Stack

- Go
- Gin
- MySQL
- `database/sql` with `go-sql-driver/mysql`

## Data Tables

- `projects`: project name, description, and fixed-project flag.
- `daily_tasks`: tasks assigned to a date and project, including estimated
  minutes and timer status.
- `time_sessions`: timer sessions for daily tasks, including start/end times
  and duration in seconds.

The SQL files are in `migrations/`.

## API

The server listens on `http://localhost:8085`.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/health` | Application health check |
| GET | `/api/health/db` | Database health check |
| POST | `/api/projects` | Create a project |
| GET | `/api/projects` | List projects |
| GET | `/api/projects/:id` | Get a project |
| PUT | `/api/projects/:id` | Update a project |
| DELETE | `/api/projects/:id` | Delete a project |
| POST | `/api/daily-tasks` | Create a daily task |
| GET | `/api/daily-tasks?date=YYYY-MM-DD` | List tasks for a date |
| GET | `/api/daily-tasks/:id` | Get a daily task |
| PUT | `/api/daily-tasks/:id` | Update a daily task |
| DELETE | `/api/daily-tasks/:id` | Delete a daily task |
| POST | `/api/daily-tasks/:id/start` | Start a planned task |
| POST | `/api/daily-tasks/:id/pause` | Pause a running task |
| POST | `/api/daily-tasks/:id/resume` | Resume a paused task |
| POST | `/api/daily-tasks/:id/finish` | Finish a running or paused task |
| GET | `/api/stats/daily?date=YYYY-MM-DD` | Get daily statistics |

Updating a daily task changes `project_id`, `task_date`, `title`, and
`estimated_minutes`. Task status is controlled by the timer endpoints.

## Local Setup

Run the setup commands from the repository root. Replace
`personal_study_timer` below with the `DB_NAME` configured in `.env`.

1. Create a MySQL database.
2. Copy the environment example and update the database values:

   ```bash
   cp .env.example .env
   ```

3. Apply the migrations in order:

   ```bash
   mysql -u root -p personal_study_timer < backend-go/migrations/001_create_projects.sql
   mysql -u root -p personal_study_timer < backend-go/migrations/002_create_daily_tasks.sql
   mysql -u root -p personal_study_timer < backend-go/migrations/003_create_time_sessions.sql
   ```

4. Start the API from the backend directory:

   ```bash
   cd backend-go
   go run ./cmd/server
   ```

## Curl Examples

Create and list projects:

```bash
curl -X POST http://localhost:8085/api/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"Go Study","description":"Backend practice","is_fixed":true}'

curl http://localhost:8085/api/projects
```

Create, list, get, and update a daily task:

```bash
curl -X POST http://localhost:8085/api/daily-tasks \
  -H 'Content-Type: application/json' \
  -d '{"project_id":1,"task_date":"2026-06-07","title":"Study Gin","estimated_minutes":45}'

curl 'http://localhost:8085/api/daily-tasks?date=2026-06-07'
curl http://localhost:8085/api/daily-tasks/1

curl -X PUT http://localhost:8085/api/daily-tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"project_id":1,"task_date":"2026-06-07","title":"Study Gin routing","estimated_minutes":60}'
```

Use the timer and view daily statistics:

```bash
curl -X POST http://localhost:8085/api/daily-tasks/1/start
curl -X POST http://localhost:8085/api/daily-tasks/1/pause
curl -X POST http://localhost:8085/api/daily-tasks/1/resume
curl -X POST http://localhost:8085/api/daily-tasks/1/finish

curl 'http://localhost:8085/api/stats/daily?date=2026-06-07'
curl -X DELETE http://localhost:8085/api/daily-tasks/1
```
