@echo off
set ROOT=E:\Projects\personal-study-timer

start "personal-timer-backend" powershell -NoExit -Command "cd '%ROOT%\backend-go'; go run ./cmd/server"

timeout /t 2 /nobreak >nul

start "personal-timer-desktop" powershell -NoExit -Command "cd '%ROOT%\desktop-wails'; wails dev"

exit
