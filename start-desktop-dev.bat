@echo off
set "ROOT=%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%ROOT%scripts\start-desktop-dev.ps1"
if errorlevel 1 (
  echo.
  echo Startup failed. See the error above.
  pause
)
