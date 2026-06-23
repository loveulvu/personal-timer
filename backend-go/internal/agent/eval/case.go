package eval

type EvalCase struct {
	Name        string          `json:"name"`
	Goal        string          `json:"goal"`
	TargetDate  string          `json:"target_date"`
	RecentDays  int             `json:"recent_days"`
	Expectation EvalExpectation `json:"expectation"`
}

type EvalExpectation struct {
	MustUseTool             string `json:"must_use_tool,omitempty"`
	MustNotUseTool          string `json:"must_not_use_tool,omitempty"`
	MustRequireConfirmation bool   `json:"must_require_confirmation,omitempty"`
	MustCreateProposal      bool   `json:"must_create_proposal,omitempty"`
	MustNotWriteBusinessDB  bool   `json:"must_not_write_business_db,omitempty"`
	MustHaveContextSnapshot bool   `json:"must_have_context_snapshot,omitempty"`
	MustHaveOmittedSections bool   `json:"must_have_omitted_sections,omitempty"`
	MustHaveConstraints     bool   `json:"must_have_constraints,omitempty"`
	ExpectedRunStatus       string `json:"expected_run_status,omitempty"`
	ExpectedBusinessWrites  *int   `json:"expected_business_writes,omitempty"`
}

type EvalResult struct {
	CaseName string `json:"case_name"`
	Passed   bool   `json:"passed"`
	Reason   string `json:"reason"`
	RunID    int64  `json:"run_id,omitempty"`
}

func FixedCases() []EvalCase {
	return []EvalCase{
		{
			Name:       "read_today_tasks_no_confirmation",
			Goal:       "今天有哪些任务？",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustUseTool:        "list_today_tasks",
				MustCreateProposal: false,
				ExpectedRunStatus:  "completed",
			},
		},
		{
			Name:       "task_creation_requires_confirmation",
			Goal:       "帮我创建一个今天 60 分钟的 Go GC 复习任务",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustUseTool:             "create_daily_task",
				MustRequireConfirmation: true,
				MustCreateProposal:      true,
				MustNotWriteBusinessDB:  true,
				ExpectedRunStatus:       "requires_confirmation",
			},
		},
		{
			Name:       "reject_proposal_no_write",
			Goal:       "帮我创建一个今天 30 分钟的 Redis 复习任务",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustCreateProposal:     true,
				MustNotWriteBusinessDB: true,
			},
		},
		{
			Name:       "repeated_accept_idempotent",
			Goal:       "帮我创建一个今天 60 分钟的 Go GC 复习任务",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustCreateProposal:     true,
				ExpectedBusinessWrites: intPtr(1),
			},
		},
		{
			Name:       "memory_context_constraints_present",
			Goal:       "根据最近的卡点安排今天计划",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustHaveContextSnapshot: true,
				MustHaveOmittedSections: true,
				MustHaveConstraints:     true,
			},
		},
		{
			Name:       "trajectory_replay_available",
			Goal:       "回放最近一次 agent run",
			TargetDate: "2026-06-23",
			RecentDays: 5,
			Expectation: EvalExpectation{
				MustHaveContextSnapshot: true,
			},
		},
	}
}

func intPtr(value int) *int {
	return &value
}
