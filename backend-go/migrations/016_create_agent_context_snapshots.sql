CREATE TABLE IF NOT EXISTS agent_context_snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    run_id BIGINT NOT NULL,
    context_json JSON NOT NULL,
    token_estimate INT NULL,
    omitted_sections_json JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_agent_context_snapshots_run_id (run_id),
    CONSTRAINT fk_agent_context_snapshots_run
        FOREIGN KEY (run_id) REFERENCES agent_runs(id)
        ON DELETE CASCADE
);
