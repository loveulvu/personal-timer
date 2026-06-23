CREATE TABLE IF NOT EXISTS agent_action_proposals (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    run_id BIGINT NOT NULL,
    step_id BIGINT NULL,
    tool_name VARCHAR(128) NOT NULL,
    action_type VARCHAR(128) NOT NULL,
    payload_json JSON NOT NULL,
    risk_level VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    result_json JSON NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    decided_at DATETIME NULL,
    executed_at DATETIME NULL,
    target_entity_type VARCHAR(64) NULL,
    target_entity_id BIGINT NULL,

    INDEX idx_agent_action_proposals_run_id (run_id),
    INDEX idx_agent_action_proposals_status (status),
    INDEX idx_agent_action_proposals_tool_name (tool_name),
    CONSTRAINT fk_agent_action_proposals_run
        FOREIGN KEY (run_id) REFERENCES agent_runs(id)
        ON DELETE CASCADE
);
