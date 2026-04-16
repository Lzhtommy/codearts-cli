---
name: codearts-pipeline
version: 0.1.1
description: "CodeArts 流水线：启动流水线（RunPipeline）、停止流水线（StopPipelineRun）。当用户需要触发 CI/CD 流水线、停止正在运行的流水线实例时使用。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli pipeline --help"
---

# codearts-pipeline (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../codearts-shared/SKILL.md`](../codearts-shared/SKILL.md) 了解配置与认证。**

CodeArts 流水线模块。`--project-id` 在所有 pipeline 命令中是**必填**的（不从 config 兜底）。

## 命令

### pipeline run

启动一条流水线。

```bash
# 最简：使用流水线保存的默认参数
codearts-cli pipeline run <pipeline_id> --project-id <project_id>

# 指定源分支
codearts-cli pipeline run <pipeline_id> --project-id <project_id> \
  --sources '[{"type":"code","params":{"build_type":"branch","target_branch":"main"}}]'

# 注入自定义变量
codearts-cli pipeline run <pipeline_id> --project-id <project_id> \
  --variables '[{"name":"ENV","value":"staging"}]'

# 用文件传完整 body
codearts-cli pipeline run <pipeline_id> --project-id <project_id> --body-file run.json

# 预览请求但不发送
codearts-cli pipeline run <pipeline_id> --project-id <project_id> --dry-run
```

**API**: `POST /v5/{project_id}/api/pipelines/{pipeline_id}/run`

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--sources` / `--sources-file` | JSON 数组：源覆盖（分支、制品…） |
| `--variables` / `--variables-file` | JSON 数组：`{name, value}` 变量 |
| `--body` / `--body-file` | 完整请求体（与 sources/variables 互斥） |
| `--description` | 运行备注 |
| `--choose-job` / `--choose-stage` | 只跑指定 job/stage（可重复） |
| `--dry-run` | 预览请求 |

**返回值**：`pipeline_run_id`（后续可用于 `pipeline stop`）。

### pipeline stop

停止一个正在运行的流水线实例。

```bash
codearts-cli pipeline stop <pipeline_id> <pipeline_run_id> --project-id <project_id>
```

**API**: `POST /v5/{project_id}/api/pipelines/{pipeline_id}/pipeline-runs/{pipeline_run_id}/stop`

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--dry-run` | 预览请求 |

## 典型工作流

```bash
# 1. 触发
RUN_ID=$(codearts-cli pipeline run <pid> --project-id <proj> 2>/dev/null | jq -r '.pipeline_run_id')

# 2. 需要时停止
codearts-cli pipeline stop <pid> $RUN_ID --project-id <proj>
```

## 注意事项

- `--project-id` 对 pipeline 命令是**必填**的，不从 config.json 兜底——防止误操作到错误项目的流水线。
- 不传 `--sources` / `--variables` / `--body` 时，API 使用流水线存储的默认配置。
- `--sources` 和 `--variables` 是 JSON 数组格式，可用 `--sources-file` 从文件加载。
