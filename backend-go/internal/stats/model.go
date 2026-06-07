package stats

type DailyTaskStat struct {
	TaskID           int64  `json:"task_id"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualSeconds    int    `json:"actual_seconds"`
	ActualMinutes    int    `json:"actual_minutes"`
}

type DailyStats struct {
	Date            string          `json:"date"`
	TotalSeconds    int             `json:"total_seconds"`
	TotalMinutes    int             `json:"total_minutes"`
	CompletedCount  int             `json:"completed_count"`
	UnfinishedCount int             `json:"unfinished_count"`
	Tasks           []DailyTaskStat `json:"tasks"`
}

type WeeklyDayStat struct {
	Date            string `json:"date"`
	TotalSeconds    int    `json:"total_seconds"`
	TotalMinutes    int    `json:"total_minutes"`
	CompletedCount  int    `json:"completed_count"`
	UnfinishedCount int    `json:"unfinished_count"`
}

type WeeklyProjectStat struct {
	ProjectID      int64  `json:"project_id"`
	ProjectName    string `json:"project_name"`
	TaskCount      int    `json:"task_count"`
	CompletedCount int    `json:"completed_count"`
	TotalSeconds   int    `json:"total_seconds"`
	TotalMinutes   int    `json:"total_minutes"`
}

type WeeklyStats struct {
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	TotalSeconds    int                 `json:"total_seconds"`
	TotalMinutes    int                 `json:"total_minutes"`
	CompletedCount  int                 `json:"completed_count"`
	UnfinishedCount int                 `json:"unfinished_count"`
	Days            []WeeklyDayStat     `json:"days"`
	Projects        []WeeklyProjectStat `json:"projects"`
}
