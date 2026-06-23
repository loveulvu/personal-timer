CREATE TABLE IF NOT EXISTS agent_steps (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    run_id BIGINT NOT NULL,
    step_index INT NOT NULL,
    step_type VARCHAR(64) NOT NULL,
    tool_name VARCHAR(128) NULL,
    tool_input_json JSON NULL,
    tool_output_json JSON NULL,
    thought_summary TEXT NULL,
    status VARCHAR(32) NOT NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_agent_steps_run_id (run_id),
    CONSTRAINT fk_agent_steps_run_id
        FOREIGN KEY (run_id) REFERENCES agent_runs(id)
        ON DELETE CASCADE
);
