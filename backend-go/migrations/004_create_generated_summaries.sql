CREATE TABLE generated_summaries (
id BIGINT PRIMARY KEY AUTO_INCREMENT,
summary_type VARCHAR(20) NOT NULL,
start_date DATE NOT NULL,
end_date DATE NOT NULL,
content TEXT NOT NULL,
source_data JSON NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
INDEX idx_summary_type_created_at (summary_type, created_at),
INDEX idx_summary_date_range (start_date, end_date),
CONSTRAINT chk_generated_summaries_type
CHECK (summary_type IN ('daily', 'weekly'))
);
