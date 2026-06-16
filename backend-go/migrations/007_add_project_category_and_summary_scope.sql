ALTER TABLE projects
    ADD COLUMN category VARCHAR(32) NOT NULL DEFAULT 'study' AFTER is_fixed,
    ADD COLUMN include_in_summary BOOLEAN NOT NULL DEFAULT TRUE AFTER category;
