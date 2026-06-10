# Personal Study Timer Desktop

This is the Wails desktop client for Personal Study Timer.

Implemented:

- backend status check
- config status check
- Today task list
- create daily task
- timer start / pause / resume / finish
- Projects CRUD

Not implemented yet:

- stats page
- summaries page
- weekly stats page
- LLM test page
- settings page
- time sessions correction UI
- auto-start backend process
- tray, notifications, shortcuts
- packaging or release config

## Prerequisites

Install and verify Wails first:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

## Run on Windows

Start the backend first:

```powershell
cd E:\Projects\personal-study-timer\backend-go
go run ./cmd/server
```

Then start the desktop app:

```powershell
cd E:\Projects\personal-study-timer\desktop-wails
wails dev
```

The desktop app calls the backend through Wails Go methods. The frontend does not directly fetch the local API.

Default backend API:

```text
http://127.0.0.1:8085
```

The user must start `backend-go` manually before using the desktop app. The desktop app does not
automatically start the backend process.

The current desktop scope does not include login or multi-user permissions, Redis, MQ, monthly
statistics, or a category system.
