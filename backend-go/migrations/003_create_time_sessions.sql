CREATE TABLE time_sessions(
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    daily_task_id BIGINT ,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    duration_seconds INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_daily_task_id (daily_task_id),
    INDEX idx_started_at (started_at),
    CONSTRAINT fk_time_sessions_daily_tasks
    FOREIGN KEY (daily_task_id)
    REFERENCES daily_tasks(id)
    

)