CREATE TABLE daily_tasks(
id BIGINT PRIMARY KEY AUTO_INCREMENT,
project_id BIGINT,
task_date DATE NOT NULL,
title TEXT NOT NULL,
estimated_minutes INT NOT NULL,
status VARCHAR(20) NOT NULL DEFAULT 'planned',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
INDEX idx_task_date (task_date),
INDEX idx_project_id_task_date(project_id,task_date),
CONSTRAINT fk_daily_tasks_project
FOREIGN KEY (project_id)
REFERENCES projects(id)
ON DELETE SET NULL
ON UPDATE CASCADE

);
 