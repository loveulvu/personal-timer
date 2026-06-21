CREATE TABLE study_feedback (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    target_type VARCHAR(50) NOT NULL,
    target_id BIGINT NOT NULL,
    target_index INT NULL,
    feedback_value VARCHAR(50) NOT NULL,
    feedback_note TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_study_feedback_target (target_type, target_id, target_index),
    INDEX idx_study_feedback_created_at (created_at)
);
