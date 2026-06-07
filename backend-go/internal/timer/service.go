package timer

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartTask(taskID int64) error {
	return s.repo.StartTask(taskID)
}
func (s *Service) PauseTask(taskID int64) error {
	return s.repo.PauseTask(taskID)
}
func (r *Repository) ResumeTask(taskID int64) error {
	return s.repo.ResumeTask(taskID)
}
