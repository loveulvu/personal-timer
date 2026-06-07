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
