# Interview Notes: Agent Harness

## 30 秒介绍

这是一个本地桌面端 Personal Agent Harness Workbench，用 Go、Wails、React 和 MySQL 实现。它不只是计时器，而是把学习计划、真实专注时长、summary、memory、feedback 和 agent proposal 串成一个可审计的数据闭环。重点是 Agent Harness：上下文裁剪、工具注册、权限控制、轨迹记录和 eval/replay。

## 1 分钟介绍

项目从个人学习计时器演进成 Agent Harness Workbench。传统 LLM 应用容易把模型输出直接当结果，我这里把模型放在受控 harness 里：ContextPack 只给结构化且裁剪过的数据；Tool Registry 声明 read/write 风险；read tool 可以自动执行；write tool 只能生成 action proposal；用户 accept 后才调用真实 service。所有 run、step、context snapshot、proposal 都能查 trajectory，并且有 eval/replay 检查关键行为。

## 3 分钟介绍

系统底层是 Go + Gin + MySQL，桌面端是 Wails + React。业务数据包括 daily tasks、time sessions、summaries、study memories、memory evidence、feedback 和 action item acceptances。Agent 部分没有直接上 LangChain，而是从工程边界做起：Tool Registry 控制模型能请求什么，ContextPack 控制模型能看到什么，Permission Guard 控制模型不能直接写库，Trajectory Log 让每次 agent run 可审计，Evaluation / Replay 用固定 case 验证 read 自动执行、write 必须 proposal、reject 不写库、accept 幂等。这个项目展示的不是“模型有多聪明”，而是如何把 LLM 决策接进真实业务系统时保留控制权。

## Agent 和普通 LLM API 有什么区别？

普通 LLM API 主要是 prompt -> response。Agent Harness 多了环境状态、工具、权限、轨迹和评估。模型只能提出 tool call 或 final answer，真正执行由后端 harness 控制。

## 为什么 write tool 不能直接执行？

因为写操作会改变用户计划和历史数据。模型可能误解上下文或生成错误参数，所以 write tool 必须先变成 proposal。用户确认后，后端再调用真实 service，并且 repeated accept 要幂等。

## Context Pack 怎么裁剪？

不把全部历史塞进 prompt。按 target date、recent days、active memory、confidence、evidence、summary excerpt 来选择。被省略或截断的部分写入 `omitted_sections`，方便审计和 eval。

## Memory 怎么避免幻觉？

Memory 来自结构化 summary/source data 和 evidence，不是让模型自由编造。ContextPack 只放 active 且置信度足够的 memory，并尽量带 evidence excerpt。archived、wrong、outdated 或低置信 memory 不进入上下文或降权。

## Trajectory 有什么用？

Trajectory 能回答“这次 agent 为什么这么做”：它保存 run、context snapshot、steps、tool input/output、proposal 和错误信息。后续 UI、debug、eval、replay 都靠它。

## Evaluation 怎么做？

Phase G V1 用固定 case 检查 harness 行为：read tool 不需要确认，write tool 必须 proposal，reject 不写业务库，重复 accept 不重复创建任务，context snapshot 和 constraints 必须存在，replay 不重新调用 LLM。

## 为什么不用 LangChain？

当前问题不是缺少 agent framework，而是要控制业务边界。工具风险、用户确认、MySQL 状态、幂等写入、trajectory 和 eval 都需要自己定义。等系统真的需要复杂 agent orchestration 时再考虑框架。

## 项目还有什么不足？

- 真实 ModelClient 还没有接入。
- Eval cases 还偏少。
- Agent Console 是演示级 UI。
- 没有大规模非结构化文档，因此暂时没有 RAG。
- 截图和 README 展示还需要补齐。
