# Personal Timer Roadmap

Personal Timer 不是普通学习计时器，也不是简单 AI 总结器。它的目标是面向学生、实习准备和项目推进场景，逐步成为一个智能学习行为分析系统。

当前优先级不是“让 AI 看起来很强”，而是让系统能稳定记录事实、生成可追溯分析、沉淀可复用建议，并真正改变用户下一次学习行为。

## 当前项目定位

Personal Timer 的核心问题是：

- 用户每天计划了什么。
- 实际投入了多少时间。
- 哪些项目稳定推进。
- 哪些任务反复超时。
- 哪些技术点反复卡住。
- 明天应该怎么调整。

所以它不是聊天机器人，也不是 Agent workflow。它首先是一个结构化学习行为数据系统，其次才是 AI 总结器。

## 当前已经完成的能力

- 项目管理：创建、编辑、分类、设置是否纳入 Summary。
- 每日任务：创建、开始、暂停、继续、完成、补充完成备注。
- 原始计时：用 `time_sessions` 记录多段计时。
- 人工修正实际耗时：`actual_seconds_override` 优先于 session 聚合。
- Daily Summary Historical Context：目标日 + 最近有数据日期。
- Weekly Summary Historical Context：本周 + 上周轻量对比。
- Project Category + Summary Scope Filter：排除 life / break / unassigned 等非学习数据。
- LLM timeout / logging / retry：避免 Weekly prompt 稍大就被短 timeout 卡死，并保留上游错误线索。
- Summary Action Items V1：基于 `source_data` 确定性生成结构化行动建议。
- Empty Daily Summary Fallback：目标日无学习数据时跳过 LLM，保存 fallback content、source_data、action_items。

## 当前还不是完整智能系统的原因

现在系统能总结和建议，但还没有形成长期学习状态。

缺口：

- action_items 还没有前端展示。
- action_items 还不能被用户采纳为明日任务。
- 没有长期 memory 表，无法把多次 Summary 中稳定出现的模式沉淀出来。
- 没有 evidence 表，无法追溯 memory 的来源和置信度。
- Memory Recall V1 已接入，但还没有 memory 管理 UI、反馈闭环或 Agent workflow。
- 没有自动估时和计划风险预测。
- 没有用户反馈闭环来验证建议是否有用。

## 为什么先做 source_data / action_items / fallback，而不是直接做 Agent

Agent 需要稳定工具、稳定状态和明确目标。当前最缺的不是 Agent，而是可信数据底座。

先做 `source_data`：

- 固定 LLM 输入事实。
- 支持复现和调试。
- 防止 LLM 编造趋势。
- 后续 memory 可以从结构化事实沉淀。

先做 `action_items`：

- 把建议从 Markdown 文案变成结构化数据。
- 后续才能展示、采纳、转任务。
- 用确定性规则覆盖第一批高价值建议，减少 LLM JSON 不稳定。

先做 fallback：

- 空数据时不浪费 LLM。
- 不让模型硬写趋势。
- 保证 Summary 生成流程稳定落库。

Agent 应该在“系统已经有可操作的结构化状态和工具”之后再考虑。

## 为什么第一版 memory 用 MySQL 结构化表，而不是向量库

第一版 memory 主要是明确事实和行为模式：

- 某项目经常超时。
- 某固定项目缺勤。
- 某技术点反复出现。
- 某类任务通常需要 60 分钟而不是 30 分钟。

这些都可以用 MySQL 字段、计数、时间范围、置信度和 evidence 表表达。向量库适合语义召回，但当前最需要的是可解释、可追溯、可更新的结构化记忆。

先用 MySQL 的理由：

- 已有数据库。
- 易调试。
- 可用 SQL 查 evidence。
- 可做状态和置信度。
- 不需要引入 embedding 成本和召回不确定性。

向量库可以等到需要“跨项目、跨文本的模糊语义检索”时再加。

## summary 和 memory 的区别

Summary 是一次生成结果：

- 范围明确：某天或某周。
- 内容包含 Markdown、source_data、action_items。
- 可以看作当时的一张快照。

Memory 是跨多次记录沉淀出的长期判断：

- 需要多条 evidence 支撑。
- 有置信度和状态。
- 可被后续 Summary recall。
- 会随新数据增强或被反证。

一句话：Summary 是“这次怎么看”，Memory 是“系统长期记住了什么”。

## action_items 和 task 的区别

Action item 是建议：

- 来源于 Summary source_data。
- 有 priority、reason、suggested_project、suggested_minutes。
- 不代表用户已经承诺执行。

Task 是计划：

- 用户明确创建。
- 有日期、项目、标题、预计分钟数和状态。
- 会进入计时和统计。

后续“一键采纳”应该把 action_item 转成 task，但 V1 不自动创建，避免系统替用户做承诺。

## memory 和 evidence 的关系

Memory 不能凭空存在。每条 memory 都应该能追溯到 evidence。

Evidence 可以来自：

- `generated_summaries.source_data`
- `generated_summaries.action_items`
- `daily_tasks.finish_note`
- `daily_tasks.finish_description`
- repeated_notes
- project overrun statistics

Memory 是归纳结果，evidence 是证据链。没有 evidence 的 memory 不应该用于影响用户计划。

## 开发原则

- 少写不必要代码。
- 不为了展示 AI 而加 AI。
- 不为了展示 Agent 而引入 Agent 框架。
- 能用 MySQL 结构化字段解决，先不用向量库。
- 能用确定性规则解决，先不要交给 LLM。
- 每个新表必须解释业务必要性。
- 每个 AI 输出都应该尽量结构化、可落库、可追溯、可复用。
- 没有改变用户下一次使用行为的 AI 输出，不算核心功能。
- 不做 Monthly，直到 Daily / Weekly 真正形成闭环。

## 路线图

### V1：数据库文档化

目标：

- 梳理当前真实表结构、字段语义、读写流程和 Summary 数据流。

业务价值：

- 降低后续改动误伤概率。
- 明确哪些数据是事实、哪些是生成结果、哪些是建议。

需要改：

- 表：不改。
- API：不改。
- UI：不改。
- 文档：新增 `docs/db_design.md` 和本路线文档。

不做：

- 不新增表。
- 不改 Summary 逻辑。
- 不做 memory。

验收标准：

- 文档覆盖 `projects`、`daily_tasks`、`time_sessions`、`generated_summaries`。
- 文档解释 `source_data`、`action_items`、Scope Filter、Empty Daily Summary Fallback。

### V2：action_items 前端展示

目标：

- 在 Summary 详情页展示已保存的 action_items。

业务价值：

- 用户不用从 Markdown 里找建议。
- 为后续“一键采纳”提供入口。

需要改：

- 表：不改。
- API：已有 `action_items` 可返回，必要时补充类型。
- UI：Summary detail 下方增加行动建议列表。

不做：

- 不自动创建任务。
- 不编辑 action_items。
- 不让 LLM 重新生成 action_items。

验收标准：

- Daily / Weekly Summary 详情能显示 action_items。
- 空数组时不报错。
- 每条展示 type、priority、title、reason、suggested_project、suggested_minutes。

### V3：action_items 采纳，用户点击后创建明日任务

目标：

- 用户点击 action_item，把建议转换为明日 `daily_tasks`。

业务价值：

- Summary 建议真正改变下一次使用行为。
- 固定项目缺失、项目拆分、技术复盘能进入计划。

需要改：

- 表：可先不改；必要时给 action_items 加 adopted 状态需要新表或 JSON 更新策略。
- API：新增 adopt action item endpoint，或前端调用现有 create task。
- UI：action_item 上增加“采纳”按钮。

不做：

- 不自动采纳。
- 不批量盲目创建。
- 不把所有建议都变成任务。

验收标准：

- 用户点击后创建明日任务。
- project、title、estimated_minutes 正确带入。
- 重复点击有防重策略。

### V4：study_memories

目标：

- 建立长期结构化学习记忆表。

业务价值：

- 系统能记住跨天/跨周稳定出现的学习模式。
- 后续 Summary 不只看最近窗口，也能参考长期模式。

草案字段：

```text
id BIGINT PRIMARY KEY
memory_type VARCHAR(32)
scope_type VARCHAR(32)
project_id BIGINT NULL
title VARCHAR(255)
content TEXT
structured_data JSON NULL
confidence DECIMAL(5,4)
support_count INT
contradiction_count INT
first_seen_at DATETIME
last_seen_at DATETIME
status VARCHAR(20)
created_at DATETIME
updated_at DATETIME
```

第一版 `memory_type`：

- `estimate_bias`：某项目或任务类型经常低估/高估。
- `consistency`：固定项目连续性问题。
- `start_time_pattern`：常见开始时间偏晚或稳定。
- `focus_topic`：反复出现的技术点。
- `scope_cleanup`：未绑定项目或分类问题。

字段说明：

- `scope_type`：`global` / `project` / `task_pattern`。
- `project_id`：项目级 memory 才填写。
- `structured_data`：保存 overrun_rate、avg_actual_minutes、keywords 等机器可读信息。
- `confidence`：当前可信度。
- `support_count`：支持证据数量。
- `contradiction_count`：反证数量。
- `status`：`active` / `archived` / `superseded`。

需要改：

- 表：新增 `study_memories`。
- API：内部查询和写入接口。
- UI：第一版可不展示。

不做：

- 不做向量库。
- 不做 Agent。
- 不让 LLM 自由写长期记忆。

验收标准：

- 能从 Summary/action_items 中沉淀 memory。
- memory 有类型、置信度、支持次数。
- inactive memory 不参与 recall。

### V5：study_memory_evidence

目标：

- 为每条 memory 建立证据链。

业务价值：

- 用户和开发者能追溯“系统为什么这么认为”。
- 支持 memory 被增强或反证。

草案字段：

```text
id BIGINT PRIMARY KEY
memory_id BIGINT NOT NULL
source_type VARCHAR(32)
source_id BIGINT
evidence_date DATE
excerpt TEXT
weight DECIMAL(5,4)
created_at DATETIME
```

`source_type` 第一版：

- `summary`
- `action_item`
- `daily_task`
- `repeated_note`

为什么 evidence 必须存在：

- memory 是归纳，不是原始事实。
- 没有 evidence 就无法判断是否可信。
- 后续用户质疑时可以回看来源。
- 新数据出现时可以增加 support 或 contradiction。

需要改：

- 表：新增 `study_memory_evidence`。
- API：memory 写入时同时写 evidence。
- UI：可后置。

不做：

- 不做复杂证据评分模型。
- 不做语义检索。

验收标准：

- 每条 memory 至少有一条 evidence。
- evidence 能回到 source summary/task。
- 删除或归档 memory 不丢 evidence。

### V6：memory recall

目标：

- 在 Daily / Weekly Summary 构造 source_data 时召回少量相关 memory。

业务价值：

- Summary 能参考长期模式，而不是只看最近几天。
- 例如“后端学习过去三周一直低估 30%-50%”。

memory 如何沉淀：

- 从 `source_data.project_breakdown` 发现长期 overrun。
- 从 `action_items` 发现重复建议。
- 从 `repeated_notes` 发现技术关键词反复出现。
- 从 Empty Daily fallback 发现固定项目缺失。

memory 如何召回：

- Daily：按目标日项目、固定项目、近期 repeated_notes 查 active memory。
- Weekly：按本周项目、overrun、start_time_patterns 查 active memory。
- 第一版只用 SQL 条件和少量 limit。

需要改：

- 表：不新增，使用 V4/V5。
- API：Summary source_data 增加 `memories` 或 `memory_context`。
- UI：不改。

不做：

- 不做 embedding。
- 不做自由文本海量召回。
- 不让 LLM 决定召回集合。

验收标准：

- source_data 包含少量可追溯 memory。
- Prompt 明确 memory 只能作为历史模式参考。
- memory 不足时 Summary 不编造长期趋势。

### V7：自动估时

目标：

- 创建任务时提示 estimated_minutes 是否过于乐观。
- 后端提供确定性估时预览，不自动改用户输入。

业务价值：

- 帮用户更现实地安排当天任务。
- 减少项目推进类任务长期超时。

第一版规则：

- 接口：`POST /api/tasks/estimate-preview`。
- 输入：`project_id`、`title`、`estimated_minutes`。
- 数据来源：最近 20 条同项目 `completed` 任务。
- 实际耗时：优先 `actual_seconds_override > 0`，否则聚合 `time_sessions.duration_seconds`；转换分钟时使用向下取整。
- 样本少于 3 条时返回 `risk_level = insufficient_data`。
- 样本充足时计算 `avg_estimated_minutes`、`avg_actual_minutes`、`overrun_rate`。
- `suggested_minutes` 使用平均实际分钟数向上取整到 5 分钟单位，且不低于用户输入。
- 风险等级：输入估时低于平均实际耗时 70% 为 high，低于 90% 为 medium，否则 low。
- 平均实际耗时达到 90 分钟时返回 `split_recommended = true`。
- reason 由后端字符串规则生成，不调用 LLM。

需要改：

- 表：不改。
- API：新增 estimate preview endpoint。
- UI：创建任务表单旁展示建议。

不做：

- 不自动修改用户输入。
- 不引入 NLP 依赖。
- 不做复杂相似度模型。
- 不按标题语义匹配。
- 不接 LLM / Agent / RAG / 向量库。

验收标准：

- 同项目历史充足时返回 avg_actual_minutes 和 overrun_rate。
- 高风险估时给出拆分或调高建议。
- 历史不足时明确返回 insufficient_data。

### V8：今日计划风险预测

目标：

- 用户制定今日计划时，提示计划总量是否明显超过近期能力。
- 事前判断当天计划是否过载，只做风险提示，不自动修改任务。

业务价值：

- 防止当天计划过载。
- 帮用户提前拆任务或减少任务数。

第一版规则：

```text
planned_total_minutes / recent_avg_actual_minutes > 1.4 => high risk
planned_total_minutes / recent_avg_actual_minutes > 1.2 => medium risk
otherwise => low risk
```

比较数据：

- 今日 `include_in_summary=true` 项目的计划总分钟数。
- 目标日期之前最近 5 个 active days 的平均实际学习分钟数。

数据来源：

- `daily_tasks.estimated_minutes` 计算今日计划总量。
- `daily_tasks.actual_seconds_override > 0` 优先，否则聚合 `time_sessions.duration_seconds`。
- `projects.include_in_summary = true` 才纳入学习计划风险。
- 最近 active days 不包含目标日期当天。

需要改：

- 表：不改。
- API：新增 `GET /api/plans/risk?date=YYYY-MM-DD`。
- UI：今日任务页展示 risk badge 和简短原因。

不做：

- 不阻止用户创建任务。
- 不做复杂日历预测。
- 不把 life/break 项目计入学习能力。
- 不调用 LLM。
- 不考虑任务难度。
- 不考虑用户当天可用时间。
- 不做自动调度。

验收标准：

- 返回 risk level、planned_total_minutes、recent_avg_minutes。
- include_in_summary=false 的任务不进入学习计划风险。
- 数据不足时返回 insufficient_data。

### V9：用户反馈闭环

目标：

- 用户可以反馈 Summary 是否准确、action_items 是否有用、memory 是否正确。

业务价值：

- 系统不只生成建议，还能学习哪些建议真正有效。
- 降低重复无效提醒。

需要改：

- 表：新增 `study_feedback`。
- API：`POST /api/feedback`。
- UI：Summary 页面提供总结准确度反馈，action item 提供有用/没用反馈。
- Memory：第一版只提供后端和 Wails API，不做完整管理 UI。

不做：

- 不调用 LLM。
- 不让 LLM 自动改写 memory。
- summary/action_item feedback 第一版只记录，不自动影响 LLM。
- memory feedback 只调整 confidence / support_count / contradiction_count / archived。
- 不做复杂评分系统。
- 不训练模型。
- 不做社交化统计。

验收标准：

- 用户能标记 Summary 准确 / 部分准确 / 不准确。
- 用户能标记 action_item 是否有用。
- 后端保存反馈。
- memory feedback 能轻量调整 confidence 和计数。

### V10：需要时再考虑向量库 / Agent workflow

目标：

- 只在结构化 SQL recall 不够时，引入更强的语义召回或 Agent workflow。

业务价值：

- 处理跨项目、跨长文本、模糊语义相似的问题。
- 自动执行多步骤学习规划或复盘工作流。

需要改：

- 表：可能新增 embedding metadata，但不是第一选择。
- API：增加 recall / workflow 编排。
- UI：看具体场景。

不做：

- 不为了“有 Agent”而加 Agent。
- 不替代已有确定性规则。
- 不把数据库事实变成不可追溯黑盒。

验收标准：

- 有明确 SQL 结构化方案解决不了的问题。
- 有评估指标证明向量召回或 Agent 带来价值。
- 所有 AI 决策仍能追溯到 evidence。

## memory 设计草案

### study_memories

作用：

- 保存长期学习行为模式，不保存单次 Summary 文案。

字段：

- `id`
- `memory_type`
- `scope_type`
- `project_id`
- `title`
- `content`
- `structured_data`
- `confidence`
- `support_count`
- `contradiction_count`
- `first_seen_at`
- `last_seen_at`
- `status`
- `created_at`
- `updated_at`

第一版 memory_type：

- `estimate_bias`
- `consistency`
- `start_time_pattern`
- `focus_topic`
- `cleanup_pattern`

### study_memory_evidence

作用：

- 保存 memory 的证据来源。

字段：

- `id`
- `memory_id`
- `source_type`
- `source_id`
- `evidence_date`
- `excerpt`
- `weight`
- `created_at`

### memory 如何沉淀

1. Summary 生成后读取 `source_data` 和 `action_items`。
2. 对确定性信号做规则判断：
   - 同项目连续超时。
   - 固定项目多次缺失。
   - repeated_notes 多次出现 Go 关键词。
   - unassigned 多次出现。
3. 若已有类似 active memory，增加 support_count 和 last_seen_at。
4. 若新证据反向，增加 contradiction_count。
5. confidence 达到阈值后参与 recall。

### memory 如何在 Daily Summary 里召回

1. 构造 Daily source_data。
2. 从当日项目、recent_context 项目和 repeated_notes 中提取查询条件。
3. SQL 查询 active memory，限制数量。
4. 把 memory 摘要和 evidence id 放入 `source_data.memory_context`。
5. Prompt 要求 LLM 只能基于 memory_context 说明长期模式。

### 为什么第一版不需要向量库

第一版 memory 查询条件是结构化的：

- project_id
- memory_type
- status
- confidence
- last_seen_at
- repeated keyword

MySQL 足够。向量库等到需要模糊语义检索时再引入。

## 自动估时设计草案

触发点：

- 用户创建或编辑任务时。

输入：

- project_id
- title
- estimated_minutes

规则：

1. 查询最近 20 次同项目 `completed` 任务。
2. 使用 `actual_seconds_override > 0 ? actual_seconds_override : session_seconds` 得到 actual seconds。
3. 使用向下取整得到 `actual_minutes`。
4. 计算 `avg_estimated_minutes`、`avg_actual_minutes`。
5. 计算 `overrun_rate = (avg_actual_minutes - avg_estimated_minutes) / avg_estimated_minutes`。
6. 样本少于 3 条时返回 `insufficient_data`。
7. 如果当前估时明显低于历史实际均值，返回调高估时建议。
8. 如果 `avg_actual_minutes >= 90`，返回拆分建议。

职责边界：

- 后端负责统计和规则判断。
- 第一版不调用 LLM。
- 用户最终决定 estimated_minutes。

## 风险预测设计草案

触发点：

- 用户制定今日计划时。

输入：

- 今日 include_in_summary=true 的计划总分钟数。
- 最近 7 / 14 天平均有效学习分钟数。

规则：

```text
planned_total_minutes > recent_avg_minutes * 1.4 => high
planned_total_minutes > recent_avg_minutes * 1.2 => medium
otherwise => low
```

输出：

- risk_level
- planned_total_minutes
- recent_avg_minutes
- reason
- suggested_adjustment

职责边界：

- 后端负责计算。
- LLM 可解释风险，但不能改任务。
- UI 只提示，不阻止用户保存。

## 当前明确不做

- Monthly Summary。
- memory / evidence migration。
- Agent。
- RAG。
- 向量库。
- action_items 自动创建任务。
- 大 UI 重构。
- 新依赖。

这些不是永远不做，只是现在没有到需要它们的阶段。

## Memory Foundation V1 状态更新

`study_memories` 和 `study_memory_evidence` 已作为数据库基础设施加入项目。

当前范围：

- 已有 migrations。
- 已有 Go model。
- 已有 repository。
- 已有 repository 测试。

当前仍然不做：

- 不接入 Daily / Weekly Summary 生成。
- 不修改 Summary prompt。
- 不做 memory 管理 UI 或反馈闭环。
- 不让 LLM 生成 memory。
- 不做 UI。
- 不引入向量库、RAG 或 Agent。

下一步只有当 Summary/action_items 闭环稳定后，才考虑从结构化 `source_data` 和 `action_items` 中沉淀 memory。

## Memory Extraction V1 状态更新

当前已支持手动触发和 Summary 生成后自动触发的确定性 memory extraction。

已支持：

- 从 summary `source_data.repeated_notes` 提取 `repeated_blocker`。
- 从 project breakdown 的超时数据提取 `estimate_bias`。
- 从 weekly time distribution 提取 `time_pattern`。
- 为每条 memory 写入 evidence，并保证同一 summary 重复提取不重复污染数据。
- Daily / Weekly Summary 保存成功后会自动尝试提取；提取失败不影响 Summary 保存和返回。

当前仍然不做：

- 不让 LLM 提取 memory。
- 不做 UI。
- 不引入向量库、RAG 或 Agent。

## Memory Recall V1 状态更新

当前已把 MySQL 结构化 memory recall 接入 Daily / Weekly Summary。

已支持：

- 生成 Summary 前召回 active `repeated_blocker` / `estimate_bias` / `time_pattern` memories。
- 将精简后的 `relevant_memories` 写入 `source_data`。
- Prompt 中增加长期记忆参考规则。
- recall 失败只记录 warning 和日志，不影响 Summary 生成。

当前仍然不做：

- 不让 LLM 提取 memory。
- 不做向量库、RAG 或 Agent。
- 不做 memory 管理 UI 或 feedback。
