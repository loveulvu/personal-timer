package stats

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDailyStats(date string) (*DailyStats, error) {
	tasks, err := s.repo.GetDailyTaskStats(date)
	if err != nil {
		return nil, err
	}

	result := &DailyStats{
		Date:  date,
		Tasks: tasks,
	}

	for _, task := range tasks {
		result.TotalSeconds += task.ActualSeconds

		if task.Status == "completed" {
			result.CompletedCount++
		} else if task.Status != "cancelled" {
			result.UnfinishedCount++
		}
	}

	result.TotalMinutes = result.TotalSeconds / 60

	return result, nil
}

func (s *Service) GetWeeklyStats(startDate, endDate string) (*WeeklyStats, error) {
	days, err := s.repo.GetWeeklyDayStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	projects, err := s.repo.GetWeeklyProjectStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := &WeeklyStats{
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
		Projects:  projects,
	}

	for _, day := range days {
		result.TotalSeconds += day.TotalSeconds
		result.CompletedCount += day.CompletedCount
		result.UnfinishedCount += day.UnfinishedCount
	}
	result.TotalMinutes = result.TotalSeconds / 60

	return result, nil
}
