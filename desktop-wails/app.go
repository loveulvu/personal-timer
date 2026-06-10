package main

import (
	"context"

	"personal-study-timer-desktop/internal/api"
)

type App struct {
	ctx    context.Context
	client *api.Client
}

func NewApp() *App {
	return &App{
		client: api.NewClient("http://127.0.0.1:8085"),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetStartupStatus() (*api.StartupStatus, error) {
	return a.client.GetStartupStatus(a.ctx)
}

func (a *App) ListDailyTasks(date string) ([]api.DailyTask, error) {
	return a.client.ListDailyTasks(a.ctx, date)
}

func (a *App) CreateDailyTask(req api.CreateDailyTaskRequest) (*api.CreateResponse, error) {
	return a.client.CreateDailyTask(a.ctx, req)
}

func (a *App) GetProjects() ([]api.Project, error) {
	return a.client.GetProjects(a.ctx)
}

func (a *App) GetProject(id int64) (*api.Project, error) {
	return a.client.GetProject(a.ctx, id)
}

func (a *App) CreateProject(input api.ProjectInput) (*api.CreateResponse, error) {
	return a.client.CreateProject(a.ctx, input)
}

func (a *App) UpdateProject(id int64, input api.ProjectInput) error {
	return a.client.UpdateProject(a.ctx, id, input)
}

func (a *App) DeleteProject(id int64) error {
	return a.client.DeleteProject(a.ctx, id)
}

func (a *App) StartTask(id int64) error {
	return a.client.TimerAction(a.ctx, id, "start")
}

func (a *App) PauseTask(id int64) error {
	return a.client.TimerAction(a.ctx, id, "pause")
}

func (a *App) ResumeTask(id int64) error {
	return a.client.TimerAction(a.ctx, id, "resume")
}

func (a *App) FinishTask(id int64) error {
	return a.client.TimerAction(a.ctx, id, "finish")
}
