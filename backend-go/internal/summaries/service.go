package summaries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"personal/internal/llm"
	"personal/internal/stats"
)

var (
	ErrStatsQueryFailed     = errors.New("stats query failed")
	ErrLLMGenerationFailed  = errors.New("LLM generation failed")
	ErrSummaryPersistFailed = errors.New("summary persistence failed")
)

type Service struct {
	repo         *Repository
	statsService *stats.Service
	llmClient    llm.Client
}

func NewService(repo *Repository, statsService *stats.Service, llmClient llm.Client) *Service {
	return &Service{
		repo:         repo,
		statsService: statsService,
		llmClient:    llmClient,
	}
}

func (s *Service) GenerateDailySummary(ctx context.Context, date string) (*GenerateSummaryResult, error) {
	dailyStats, err := s.statsService.GetDailyStats(date)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	sourceData, err := json.Marshal(dailyStats)
	if err != nil {
		return nil, err
	}

	content, err := s.llmClient.GenerateSummary(ctx, buildDailyPrompt(string(sourceData)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
	}

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "daily",
		StartDate:   date,
		EndDate:     date,
		Content:     content,
		SourceData:  sourceData,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}

	return &GenerateSummaryResult{SummaryID: id, Content: content}, nil
}

func (s *Service) GenerateWeeklySummary(ctx context.Context, startDate, endDate string) (*GenerateSummaryResult, error) {
	weeklyStats, err := s.statsService.GetWeeklyStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	sourceData, err := json.Marshal(weeklyStats)
	if err != nil {
		return nil, err
	}

	content, err := s.llmClient.GenerateSummary(ctx, buildWeeklyPrompt(string(sourceData)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
	}

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "weekly",
		StartDate:   startDate,
		EndDate:     endDate,
		Content:     content,
		SourceData:  sourceData,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}

	return &GenerateSummaryResult{SummaryID: id, Content: content}, nil
}

func (s *Service) ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error) {
	return s.repo.ListSummaries(ctx, summaryType)
}

func (s *Service) GetSummaryByID(ctx context.Context, id int64) (*GeneratedSummary, error) {
	return s.repo.GetSummaryByID(ctx, id)
}

func buildDailyPrompt(sourceData string) string {
	return `Generate a concise daily study review from the JSON statistics below.
Focus on completed work, time allocation, unfinished tasks, estimation differences,
and practical improvements for tomorrow. Do not praise or use motivational language.

Statistics:
` + sourceData
}

func buildWeeklyPrompt(sourceData string) string {
	return `Generate a concise weekly study review from the JSON statistics below.
Focus on total time, daily fluctuations, project time distribution, completion,
main problems, and practical suggestions for next week. Do not praise or use motivational language.

Statistics:
` + sourceData
}
