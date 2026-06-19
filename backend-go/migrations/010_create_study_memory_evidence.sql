CREATE TABLE study_memory_evidence (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    memory_id BIGINT NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id BIGINT NULL,
    evidence_date DATE NOT NULL,
    excerpt TEXT NULL,
    weight DECIMAL(4,3) NOT NULL DEFAULT 1.000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_memory_evidence_memory (memory_id),
    INDEX idx_memory_evidence_source (source_type, source_id),
    INDEX idx_memory_evidence_date (evidence_date),
    CONSTRAINT fk_study_memory_evidence_memory
        FOREIGN KEY (memory_id)
        REFERENCES study_memories(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);
