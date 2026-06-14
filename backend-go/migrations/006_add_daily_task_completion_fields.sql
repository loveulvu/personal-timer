ALTER TABLE daily_tasks
    ADD COLUMN finish_note TEXT NULL AFTER status,
    ADD COLUMN finish_description TEXT NULL AFTER finish_note,
    ADD COLUMN completed_at DATETIME NULL AFTER finish_description,
    ADD COLUMN actual_seconds_override INT NULL AFTER completed_at;
