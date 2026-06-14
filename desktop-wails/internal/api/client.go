package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) GetStartupStatus(ctx context.Context) (*StartupStatus, error) {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return &StartupStatus{
			Connected: false,
			Error:     "Backend is not running. Please start backend-go first.",
		}, nil
	}

	config, err := c.GetConfigStatus(ctx)
	if err != nil && config == nil {
		return &StartupStatus{
			Connected: true,
			Version:   version,
			Error:     err.Error(),
		}, nil
	}

	status := &StartupStatus{
		Connected: true,
		Version:   version,
		Config:    config,
	}
	if err != nil {
		status.Error = err.Error()
	}
	return status, nil
}

func (c *Client) GetVersion(ctx context.Context) (*VersionInfo, error) {
	var result VersionInfo
	if err := c.doJSON(ctx, http.MethodGet, "/api/version", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetConfigStatus(ctx context.Context) (*ConfigStatus, error) {
	var result ConfigStatus
	err := c.doJSON(ctx, http.MethodGet, "/api/config/status", nil, &result)
	if err != nil {
		if result.Database != "" {
			return &result, err
		}
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListDailyTasks(ctx context.Context, date string) ([]DailyTask, error) {
	path := "/api/daily-tasks?date=" + url.QueryEscape(date)
	var result dataResponse[[]DailyTask]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) CreateDailyTask(ctx context.Context, req CreateDailyTaskRequest) (*CreateResponse, error) {
	var result CreateResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/daily-tasks", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var result dataResponse[[]Project]
	if err := c.doJSON(ctx, http.MethodGet, "/api/projects", nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) GetProject(ctx context.Context, id int64) (*Project, error) {
	var result dataResponse[Project]
	path := fmt.Sprintf("/api/projects/%d", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) CreateProject(ctx context.Context, input ProjectInput) (*CreateResponse, error) {
	var result CreateResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/projects", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateProject(ctx context.Context, id int64, input ProjectInput) error {
	path := fmt.Sprintf("/api/projects/%d", id)
	return c.doJSON(ctx, http.MethodPut, path, input, nil)
}

func (c *Client) DeleteProject(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/projects/%d", id)
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) GetDailyStats(ctx context.Context, date string) (*DailyStats, error) {
	path := "/api/stats/daily?date=" + url.QueryEscape(date)
	var result dataResponse[DailyStats]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) GetWeeklyStats(ctx context.Context, startDate, endDate string) (*WeeklyStats, error) {
	values := url.Values{}
	values.Set("start_date", startDate)
	values.Set("end_date", endDate)
	var result dataResponse[WeeklyStats]
	if err := c.doJSON(ctx, http.MethodGet, "/api/stats/weekly?"+values.Encode(), nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) GenerateDailySummary(ctx context.Context, date string) (*GenerateSummaryResult, error) {
	var result dataResponse[GenerateSummaryResult]
	req := GenerateDailySummaryRequest{Date: date}
	if err := c.doJSON(ctx, http.MethodPost, "/api/summaries/daily/generate", req, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) GenerateWeeklySummary(ctx context.Context, startDate, endDate string) (*GenerateSummaryResult, error) {
	var result dataResponse[GenerateSummaryResult]
	req := GenerateWeeklySummaryRequest{StartDate: startDate, EndDate: endDate}
	if err := c.doJSON(ctx, http.MethodPost, "/api/summaries/weekly/generate", req, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) GetSummaries(ctx context.Context, summaryType string) ([]Summary, error) {
	path := "/api/summaries"
	if summaryType != "" {
		path += "?type=" + url.QueryEscape(summaryType)
	}
	var result dataResponse[[]Summary]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) GetSummary(ctx context.Context, id int64) (*Summary, error) {
	var result dataResponse[Summary]
	path := fmt.Sprintf("/api/summaries/%d", id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) DeleteSummary(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/summaries/%d", id)
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) TestLLM(ctx context.Context) (*LLMTestResponse, error) {
	var result LLMTestResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/llm/test", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) TimerAction(ctx context.Context, id int64, action string) error {
	switch action {
	case "start", "pause", "resume":
	default:
		return fmt.Errorf("unsupported timer action: %s", action)
	}
	path := fmt.Sprintf("/api/daily-tasks/%d/%s", id, action)
	return c.doJSON(ctx, http.MethodPost, path, nil, nil)
}

func (c *Client) FinishTask(ctx context.Context, id int64, input FinishTaskRequest) error {
	path := fmt.Sprintf("/api/daily-tasks/%d/finish", id)
	return c.doJSON(ctx, http.MethodPost, path, input, nil)
}

func (c *Client) UpdateCompletedTask(ctx context.Context, id int64, input UpdateCompletedTaskRequest) error {
	path := fmt.Sprintf("/api/daily-tasks/%d/completion", id)
	return c.doJSON(ctx, http.MethodPut, path, input, nil)
}

func (c *Client) DeleteCompletedTask(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/daily-tasks/%d/completion", id)
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return errors.New("backend request timed out")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("backend request timed out")
		}
		return fmt.Errorf("Backend is not running. Please start backend-go first: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if out != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, out); err != nil {
			return err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(extractError(responseBody, resp.StatusCode))
	}

	return nil
}

func extractError(body []byte, statusCode int) string {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Error != "" {
			return payload.Error
		}
		if payload.Message != "" {
			return payload.Message
		}
	}
	return fmt.Sprintf("backend request failed with status %d", statusCode)
}
