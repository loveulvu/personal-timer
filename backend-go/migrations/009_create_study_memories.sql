CREATE TABLE study_memories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    memory_type VARCHAR(50) NOT NULL,
    scope_type VARCHAR(30) NOT NULL,
    project_id BIGINT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    structured_data JSON NULL,
    confidence DECIMAL(4,3) NOT NULL DEFAULT 0.500,
    support_count INT NOT NULL DEFAULT 0,
    contradiction_count INT NOT NULL DEFAULT 0,
    first_seen_at DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_study_memories_type_status (memory_type, status),
    INDEX idx_study_memories_scope (scope_type, project_id, status),
    INDEX idx_study_memories_last_seen (last_seen_at),
    INDEX idx_study_memories_project (project_id),
    CONSTRAINT fk_study_memories_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
);
