CREATE TABLE IF NOT EXISTS summary_action_item_acceptances (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    summary_id BIGINT NOT NULL,
    item_index INT NOT NULL,
    target_date DATE NOT NULL,
    target_task_id BIGINT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_summary_item_target_date (summary_id, item_index, target_date),
    KEY idx_summary_id (summary_id),
    KEY idx_target_task_id (target_task_id),
    KEY idx_target_date (target_date),
    CONSTRAINT fk_acceptance_summary FOREIGN KEY (summary_id) REFERENCES generated_summaries(id) ON DELETE CASCADE,
    CONSTRAINT fk_acceptance_task FOREIGN KEY (target_task_id) REFERENCES daily_tasks(id) ON DELETE SET NULL
);
