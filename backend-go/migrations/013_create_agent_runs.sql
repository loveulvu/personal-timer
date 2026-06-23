CREATE TABLE IF NOT EXISTS agent_runs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_goal TEXT NOT NULL,
    target_date DATE NOT NULL,
    status VARCHAR(32) NOT NULL,
    final_answer TEXT NULL,
    pending_actions_json JSON NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL
);
