# Personal Timer 数据库设计

本文基于当前 migrations 和 `backend-go/internal` 代码整理。当前核心表只有：

- `projects`
- `daily_tasks`
- `time_sessions`
- `generated_summaries`

没有 Monthly、memory、observation、RAG、向量库相关表。

## 设计原则

- `projects` 定义任务归属和是否进入学习分析。
- `daily_tasks` 保存用户每天计划、状态、完成备注和人工修正后的实际耗时。
- `time_sessions` 保存一次次开始/暂停/继续产生的原始计时片段。
- `generated_summaries` 保存 AI 总结结果、当次输入快照和确定性生成的行动建议。
- 实际耗时统一口径：优先使用 `daily_tasks.actual_seconds_override`；为空时使用 `time_sessions.duration_seconds` 聚合。

## projects

### 业务含义

项目是任务的归属维度，也是 Summary 统计范围的控制点。它用于区分学习、项目推进、生活、休息等不同类型，避免“吃饭”这类生活记录污染学习总结。

### 主键

- `id BIGINT PRIMARY KEY AUTO_INCREMENT`

### 外键

无外键被其它表引用：

- `daily_tasks.project_id -> projects.id`

### 核心字段

- `name VARCHAR(100) NOT NULL UNIQUE`：项目名称。
- `description TEXT`：项目说明。
- `is_fixed BOOLEAN NOT NULL DEFAULT FALSE`：是否固定项目。
- `category VARCHAR(32) NOT NULL DEFAULT 'study'`：项目分类。
- `include_in_summary BOOLEAN NOT NULL DEFAULT TRUE`：是否纳入 Daily / Weekly Summary 学习统计。
- `created_at` / `updated_at`：创建和更新时间。

### 状态字段

无生命周期状态字段。`is_fixed` 是项目属性，不是状态机。

### JSON 字段

无。

### 重点字段说明

#### `category`

第一版允许：

- `study`
- `project`
- `life`
- `break`
- `other`

它描述项目性质，用于解释 excluded 数据。例如 `life` / `break` 可以保留在系统里，但不一定进入学习分析。

#### `include_in_summary`

Summary Scope Filter 的核心开关。

- `true`：该项目任务进入 Daily / Weekly Summary 的学习统计。
- `false`：该项目任务从学习统计中排除，但可进入 `source_data.excluded`，用于说明数据范围。

旧数据默认 `true`，避免升级后历史学习项目突然从总结中消失。

### 为什么字段放在这张表

分类和是否纳入 Summary 是项目级属性，不是单次任务属性。放在 `projects` 可以保证同项目下所有任务统计口径一致，也避免每个 `daily_tasks` 重复存分类配置。

### 读取它的流程

- 创建 / 编辑 / 列出项目。
- 创建任务时选择项目。
- Daily / Weekly Summary 构造 `source_data` 时读取 `category` 和 `include_in_summary`。
- 普通统计按项目聚合时读取项目名称。

### 写入它的流程

- 创建项目。
- 更新项目。
- 删除项目。

### 典型数据流转

1. 用户创建项目，例如 `后端学习`，默认 `category='study'`、`include_in_summary=true`。
2. 用户创建项目 `吃饭`，设置 `category='life'`、`include_in_summary=false`。
3. 任务绑定项目。
4. Summary 构造时只把 `include_in_summary=true` 的项目计入学习总时长。
5. 被排除项目进入 `source_data.excluded`，供 LLM 说明数据范围，不参与效率分析。

## daily_tasks

### 业务含义

`daily_tasks` 是每天计划和完成结果的主表。它保存任务标题、所属项目、预计时长、当前状态、完成备注和人工实际耗时修正。

### 主键

- `id BIGINT PRIMARY KEY AUTO_INCREMENT`

### 外键

- `project_id BIGINT NULL`
- `FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL ON UPDATE CASCADE`

项目删除后任务保留，`project_id` 置空。未绑定项目默认不纳入 Summary 学习统计。

### 核心字段

- `project_id`：任务所属项目，可为空。
- `task_date DATE NOT NULL`：任务日期。
- `title TEXT NOT NULL`：任务标题。
- `estimated_minutes INT NOT NULL`：用户计划投入分钟数。
- `status VARCHAR(20) NOT NULL DEFAULT 'planned'`：任务状态。
- `finish_note TEXT NULL`：完成时的短备注。
- `finish_description TEXT NULL`：完成时的较长描述。
- `completed_at DATETIME NULL`：完成时间。
- `actual_seconds_override INT NULL`：人工修正后的实际秒数。
- `created_at` / `updated_at`：创建和更新时间。

### 索引

- `idx_task_date (task_date)`
- `idx_project_id_task_date (project_id, task_date)`

### 状态字段

`status` 当前代码使用：

- `planned`：已计划，未开始。
- `running`：正在计时。
- `paused`：暂停中。
- `completed`：已完成。
- `cancelled`：前端类型中保留，但当前 timer 流程主要围绕 planned/running/paused/completed。

状态流转：

```text
planned -> running -> paused -> running -> completed
planned -> running -> completed
```

### JSON 字段

无。

### 重点字段说明

#### `estimated_minutes`

用户创建任务时的计划时长。Summary 用它和实际时长对比，计算 overrun、overrun_rate，并为后续自动估时提供基础。

#### `actual_seconds_override`

完成后允许用户修正实际耗时。所有涉及实际时长的统计应优先使用它：

```text
actual_seconds = actual_seconds_override ?? SUM(time_sessions.duration_seconds)
```

这能处理忘记暂停、计时器误开、补录等真实场景。

#### `finish_note` / `finish_description`

完成任务时写入。Summary 会从这两个字段提取 repeated_notes，用于发现反复出现的技术点、阻塞点和 action_items 的 focus_topic。

### 为什么字段放在这张表

预计时长、状态、完成备注、完成时间和人工修正都是“任务级结果”，不是项目级配置，也不是单次计时片段。放在 `daily_tasks` 可以稳定描述一次任务从计划到完成的完整业务状态。

### 读取它的流程

- 按日期列出任务。
- 查看单个任务。
- 开始 / 暂停 / 继续 / 完成任务前检查状态。
- Daily / Weekly 普通统计。
- Daily / Weekly Summary 构造 `source_data`。
- action_items 确定性生成。

### 写入它的流程

- 创建任务。
- 更新任务。
- 开始任务：写 `status='running'`。
- 暂停任务：写 `status='paused'`。
- 继续任务：写 `status='running'`。
- 完成任务：写 `status='completed'`、`finish_note`、`finish_description`、`completed_at`。
- 更新已完成任务：写完成备注和 `actual_seconds_override`。
- 删除已完成任务。

### 典型数据流转

1. 用户创建今日任务，写入 `planned`。
2. 用户开始任务，状态变为 `running`，同时创建一条 `time_sessions`。
3. 用户暂停任务，状态变为 `paused`，当前 session 写入 `ended_at` 和 `duration_seconds`。
4. 用户继续任务，状态回到 `running`，新建一条 session。
5. 用户完成任务，状态变为 `completed`，写入完成备注。
6. Summary 使用任务、session、项目配置生成结构化分析。

## time_sessions

### 业务含义

`time_sessions` 是原始计时片段表。一次开始到暂停或完成，就是一条 session。一个任务可以有多条 session。

### 主键

- `id BIGINT PRIMARY KEY AUTO_INCREMENT`

### 外键

- `daily_task_id BIGINT NOT NULL`
- `FOREIGN KEY (daily_task_id) REFERENCES daily_tasks(id) ON DELETE CASCADE ON UPDATE CASCADE`

任务删除后，对应 session 自动删除。

### 核心字段

- `daily_task_id`：所属任务。
- `started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`：开始时间。
- `ended_at DATETIME NULL`：结束时间。为空表示当前 session 仍在运行。
- `duration_seconds INT NOT NULL DEFAULT 0`：本 session 持续秒数。
- `created_at`：创建时间。

### 索引

- `idx_daily_task_id (daily_task_id)`
- `idx_started_at (started_at)`

### 状态字段

无独立状态字段。通过 `ended_at IS NULL` 判断 session 是否运行中。

### JSON 字段

无。

### 重点字段说明

#### `started_at`

用于：

- 计算 session duration。
- 判断今日 `first_start_time`。
- Summary 的 time_distribution。
- Weekly 的 start_time_patterns。

#### `ended_at`

为空代表正在运行。暂停或完成时写入当前时间。

#### `duration_seconds`

暂停或完成时由 `ended_at - started_at` 计算写入。统计时被聚合到任务实际耗时，但如果任务有 `actual_seconds_override`，则 override 优先。

### 为什么字段放在这张表

一个任务可能多次开始、暂停、继续。把每段时间拆成 session 可以保留原始时间分布，也能支持 first_start_time、时间段分析、后续复盘。

### 读取它的流程

- 列出任务时聚合实际耗时和当前运行 session。
- 暂停 / 完成时找到当前未结束 session。
- 继续时不读旧 session，只新建 session。
- Daily / Weekly Summary 统计时间分布和开始时间模式。

### 写入它的流程

- 开始任务：插入一条 session。
- 继续任务：插入一条 session。
- 暂停任务：更新当前 session 的 `ended_at` 和 `duration_seconds`。
- 完成 running 任务：先关闭当前 session。
- 删除已完成任务：删除关联 session。

### 典型数据流转

```text
StartTask:
  daily_tasks.status planned -> running
  INSERT time_sessions(daily_task_id)

PauseTask:
  找 ended_at IS NULL 的 session
  UPDATE ended_at, duration_seconds
  daily_tasks.status running -> paused

ResumeTask:
  daily_tasks.status paused -> running
  INSERT 新 session

FinishTask:
  如果 running，先关闭当前 session
  daily_tasks.status -> completed
```

## generated_summaries

### 业务含义

保存 Daily / Weekly Summary 的结果和生成依据。它不是记忆表，只保存某一次总结的内容、输入快照和结构化 action_items。

### 主键

- `id BIGINT PRIMARY KEY AUTO_INCREMENT`

### 外键

无。

Summary 是生成快照，不直接外键绑定任务。这样历史 summary 不会因任务后来修改而失去当时的输入上下文。

### 核心字段

- `summary_type VARCHAR(20) NOT NULL`：`daily` 或 `weekly`。
- `start_date DATE NOT NULL`：总结范围开始日期。
- `end_date DATE NOT NULL`：总结范围结束日期。
- `content TEXT NOT NULL`：LLM 或 fallback 生成的 Markdown 总结。
- `source_data JSON NULL`：本次喂给 LLM 或 fallback 使用的结构化上下文。
- `action_items JSON NULL`：基于 `source_data` 确定性生成的行动建议。
- `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`

### 约束和索引

- `CHECK (summary_type IN ('daily', 'weekly'))`
- `UNIQUE KEY uq_generated_summaries_type_range (summary_type, start_date, end_date)`
- `idx_summary_type_created_at (summary_type, created_at)`
- `idx_summary_date_range (start_date, end_date)`

### 状态字段

无。是否已生成由唯一键和记录存在性表示。

### JSON 字段含义

#### `source_data`

保存“本次总结的输入事实”，用于调试、复现、回溯和后续沉淀 memory。

Daily 关键结构：

- `summary_type`
- `target_date`
- `data_quality`
- `today`
- `recent_context`
- `excluded`
- `warnings`

Weekly 关键结构：

- `summary_type`
- `week_start`
- `week_end`
- `data_quality`
- `week`
- `previous_week_comparison`
- `excluded`
- `warnings`

`source_data` 的意义是把 LLM 输入固定下来。即使任务后续修改，也知道当时 Summary 为什么这么写。

#### `action_items`

保存结构化行动建议，不从 Markdown content 解析，而是基于 `source_data` 确定性生成。当前字段包含：

- `type`
- `priority`
- `title`
- `reason`
- `suggested_project`
- `suggested_minutes`
- `source`

第一版类型：

- `schedule`
- `consistency`
- `estimation`
- `split_task`
- `focus_topic`
- `cleanup`

### 为什么字段放在这张表

`content`、`source_data`、`action_items` 都属于“某一次 Summary 生成结果”。放在同一表可以保证查询 summary detail 时拿到完整上下文，避免 action_items 与 content 版本错位。

### 读取它的流程

- Summary 列表。
- Summary 详情。
- 后续 action_items 前端展示。
- 后续 memory 沉淀时读取历史 source_data / action_items。

### 写入它的流程

- GenerateDailySummary。
- GenerateWeeklySummary。
- Empty Daily Summary Fallback。

### 典型数据流转

1. 后端构造 `source_data`。
2. 如果 Daily 当天没有纳入学习统计的数据，跳过 LLM，生成 fallback content。
3. 否则把 `source_data` 放进 prompt 请求 LLM。
4. 基于同一份 `source_data` 确定性生成 `action_items`。
5. 写入 `generated_summaries(content, source_data, action_items)`。

## 核心业务流程

### 1. 创建任务流程

1. 前端提交 `project_id`、`task_date`、`title`、`estimated_minutes`。
2. 后端写入 `daily_tasks`。
3. 默认 `status='planned'`。
4. 不创建 `time_sessions`。

读取：

- 可选读取 project 是否存在，当前 repository 主要负责插入任务。

写入：

- `daily_tasks`

### 2. 开始任务流程

1. 事务内锁定 `daily_tasks`。
2. 要求 `status='planned'`。
3. 插入 `time_sessions(daily_task_id)`，`started_at` 默认当前时间。
4. 更新任务状态为 `running`。
5. 提交事务。

读取：

- `daily_tasks.status`

写入：

- `time_sessions`
- `daily_tasks.status`

### 3. 暂停任务流程

1. 事务内锁定任务。
2. 要求 `status='running'`。
3. 查找该任务最新一条 `ended_at IS NULL` session。
4. 计算 `duration_seconds = now - started_at`。
5. 写入 `ended_at` 和 `duration_seconds`。
6. 更新任务状态为 `paused`。

读取：

- `daily_tasks.status`
- 当前 running `time_sessions`

写入：

- `time_sessions.ended_at`
- `time_sessions.duration_seconds`
- `daily_tasks.status`

### 4. 继续任务流程

1. 事务内锁定任务。
2. 要求 `status='paused'`。
3. 插入新的 `time_sessions`。
4. 更新任务状态为 `running`。

读取：

- `daily_tasks.status`

写入：

- `time_sessions`
- `daily_tasks.status`

### 5. 完成任务流程

1. 事务内锁定任务。
2. 要求 `status IN ('running', 'paused')`。
3. 如果 running，先关闭当前 session。
4. 更新任务：
   - `status='completed'`
   - `finish_note`
   - `finish_description`
   - `completed_at`
5. 后续用户可更新完成备注和 `actual_seconds_override`。

读取：

- `daily_tasks.status`
- running session

写入：

- `time_sessions`，仅 running 时关闭当前 session。
- `daily_tasks.status`
- `daily_tasks.finish_note`
- `daily_tasks.finish_description`
- `daily_tasks.completed_at`

### 6. 生成 Daily Summary 流程

1. 检查 `generated_summaries` 是否已有同一天 daily summary。
2. 查询目标日期之前最近 `DailyRecentActiveDaysLimit = 5` 个有数据日期。
3. 查询目标日期和历史日期窗口内的 `daily_tasks`。
4. 查询同窗口内的 `time_sessions`。
5. 按 `projects.include_in_summary` 和是否有项目绑定拆分 included / excluded。
6. 构造 `DailySummarySourceData`：
   - `today`
   - `recent_context`
   - `data_quality`
   - `excluded`
   - `warnings`
7. 如果目标日 included 数据为空，走 Empty Daily Summary Fallback。
8. 否则把 `source_data` 放进 Daily prompt 调用 LLM。
9. 基于同一份 `source_data` 生成 `action_items`。
10. 写入 `generated_summaries`。

### 7. 生成 Weekly Summary 流程

1. 检查 `generated_summaries` 是否已有同范围 weekly summary。
2. 根据入参 `start_date` / `end_date` 生成本周日期列表。
3. 查询本周 tasks / sessions。
4. 查询上一周 7 天 tasks / sessions。
5. 按 Summary Scope Filter 拆分 included / excluded。
6. 构造 `WeeklySummarySourceData`：
   - `week.daily_totals`
   - `week.project_breakdown`
   - `week.time_distribution`
   - `week.start_time_patterns`
   - `week.repeated_notes`
   - `previous_week_comparison`
   - `excluded`
7. 把 `source_data` 放进 Weekly prompt 调用 LLM。
8. 基于同一份 `source_data` 生成 `action_items`。
9. 写入 `generated_summaries`。

### 8. 生成 action_items 流程

action_items 不解析 Markdown content。

Daily：

- 缺少固定项目 `算法练习` / `背单词`：生成 `schedule high`。
- `项目推进` 超时超过阈值：生成 `split_task`。
- repeated_notes 出现 Go 高频关键词：生成 `focus_topic`。
- 存在未绑定任务：生成 `cleanup`。

Weekly：

- 固定项目活跃天数少于本周有数据天数：生成 `consistency high`。
- `项目推进` 超时或 overrun_rate 高：生成 `split_task high`。
- `后端学习` 活跃天数少：生成 `schedule medium`。
- repeated_notes 出现 Go 高频关键词：生成 `focus_topic`。
- 存在未绑定任务：生成 `cleanup`。

最后去重、按优先级排序、最多保留 5 条。

### 9. Empty Daily Summary Fallback 流程

触发条件：

```text
today.total_focus_minutes == 0
today.task_count == 0
len(today.project_breakdown) == 0
```

行为：

1. 已经构造并保存 `source_data`。
2. 已经基于 `source_data` 生成 action_items。
3. 不调用 LLM。
4. 用固定 Markdown 模板生成 Daily Summary content。
5. 如果有 excluded 数据，说明存在被排除的非学习记录。
6. 写入 `generated_summaries`。

意义：

- 避免空数据 prompt 让 LLM 编造趋势。
- 避免浪费 LLM 请求。
- 保留 source_data，便于调试为什么当天是空总结。

### 10. Project Scope Filter 如何影响 summary 统计

过滤规则：

- `project_id IS NULL`：不纳入学习统计，计入 unassigned excluded。
- `projects.include_in_summary = false`：不纳入学习统计，计入 excluded_projects。
- `include_in_summary = true`：进入学习统计。

被排除的数据不计入：

- `total_focus_minutes`
- `completed_tasks`
- `task_count`
- `project_breakdown`
- `time_distribution`
- `project_patterns`
- `start_time_patterns`
- `repeated_notes`

但会进入：

- `source_data.excluded.excluded_task_count`
- `source_data.excluded.excluded_total_minutes`
- `source_data.excluded.excluded_projects`
- `source_data.excluded.unassigned_task_count`
- `source_data.warnings`

这样 LLM 可以知道“生活类/未绑定数据已排除”，但不会把它当学习投入或学习效率问题分析。

## 实际耗时统一口径

所有统计应遵守：

```text
actual_seconds = daily_tasks.actual_seconds_override
  if actual_seconds_override IS NOT NULL
  else SUM(time_sessions.duration_seconds)
```

原因：

- `time_sessions` 是原始计时。
- `actual_seconds_override` 是用户确认后的修正值。
- Summary、stats、project breakdown 和 overrun 判断必须优先相信用户修正。

## 当前没有的表

当前没有：

- `monthly_summaries`
- `observations`
- `agents`
- embedding / vector 表

后续如果新增表，必须说明业务必要性、读写流程和是否能被现有 JSON 快照替代。

## Memory Foundation V1

当前已新增长期记忆基础表，但尚未接入 Summary 生成流程，也没有 memory recall、Agent、RAG 或向量库。

### study_memories

业务含义：保存跨天/跨周沉淀出的学习行为模式，例如估时偏差、时间模式、项目推进模式、重复阻塞和建议模式。

核心字段：

- `memory_type`：`time_pattern` / `estimate_bias` / `project_pattern` / `repeated_blocker` / `suggestion_pattern`
- `scope_type`：`global` / `project` / `topic`
- `project_id`：可选项目范围，引用 `projects(id)`，项目删除时置空。
- `structured_data`：JSON，保存机器可读的统计值。
- `confidence`、`support_count`、`contradiction_count`：置信度和证据计数。
- `first_seen_at`、`last_seen_at`：首次和最近观察时间。
- `status`：`active` / `archived`

当前读写者：仅 `backend-go/internal/memories` repository。Summary、action_items、任务流程暂不读写。

### study_memory_evidence

业务含义：保存 memory 的证据链。

核心字段：

- `memory_id`：引用 `study_memories(id)`，memory 删除时级联删除。
- `source_type`：`daily_summary` / `weekly_summary` / `daily_task` / `finish_note` / `action_item` / `manual`
- `source_id`：来源记录 id，可为空。
- `evidence_date`：证据日期。
- `excerpt`：证据摘录。
- `weight`：证据权重。

当前不会影响 Daily / Weekly Summary 生成；后续 memory recall 需要单独设计和接入。
