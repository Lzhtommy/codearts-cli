---
name: codearts-pipeline
version: 0.1.1
description: "CodeArts 流水线：列出流水线（ListPipelines）、启动流水线（RunPipeline）、停止流水线（StopPipelineRun）、查询流水线运行详情（ShowPipelineRunDetail）。当用户需要查看、触发、停止或查询 CI/CD 流水线状态时使用。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli pipeline --help"
---

# codearts-pipeline (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../codearts-shared/SKILL.md`](../codearts-shared/SKILL.md) 了解配置与认证。**

CodeArts 流水线模块。`--project-id` 在所有 pipeline 命令中是**必填**的（不从 config 兜底）。

### 从 git remote 自动提取 project-id

当用户在一个 CodeArts Repo 克隆的仓库目录下操作时，可以从 `git remote -v` 的 URL 中提取 project-id，避免手动输入：

```
git@codehub-cn-south-1.devcloud.huaweicloud.com:759278abbfb14b098eeddc548741f38b/nest-app-agent.git
                                                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                                 这就是 project-id
```

**提取方式**（Agent 应自动执行）：

```bash
PROJECT_ID=$(git remote -v | grep codehub | head -1 | sed 's/.*:\([a-f0-9]\{32\}\)\/.*/\1/')
```

提取后直接传给 `--project-id $PROJECT_ID`。如果 `git remote -v` 中没有 `codehub` 开头的 URL，则需要用户手动提供。

## 命令

### pipeline list

列出项目下的流水线。`--project-id` 同时作为 URL path 参数和 body 的 `project_id` / `project_ids`。

```bash
# 列出所有流水线
codearts-cli pipeline list --project-id <project_id>

# 按名称过滤
codearts-cli pipeline list --project-id <project_id> --name "deploy"

# 分页
codearts-cli pipeline list --project-id <project_id> --offset 0 --limit 20
```

**API 参考**: [ListPipelines](https://support.huaweicloud.com/api-pipeline/ListPipelines.html)

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--name` | 按流水线名称模糊匹配 |
| `--status` | 按状态过滤（可重复）：`COMPLETED` / `RUNNING` / `FAILED` / `CANCELED` / `PAUSED` / `SUSPEND` / `IGNORED` |
| `--creator-id` / `--executor-id` | 按创建人/执行人 user_id 过滤（可重复） |
| `--start-time` / `--end-time` | 时间范围过滤 |
| `--offset` / `--limit` | 分页 |
| `--sort-key` | 排序字段：`name` / `create_time` / `update_time` |
| `--sort-dir` | 排序方向：`asc` / `desc` |
| `--dry-run` | 预览请求 |

**返回值**：`pipelines` 数组，每条包含 `pipeline_id`、`name`、`latest_run`（含 `pipeline_run_id`、`status`）。

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

**API 参考**: [RunPipeline](https://support.huaweicloud.com/api-pipeline/RunPipeline.html)

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

**API 参考**: [StopPipelineRun](https://support.huaweicloud.com/api-pipeline/StopPipelineRun.html)

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--dry-run` | 预览请求 |

### pipeline status

查询流水线运行实例详情（阶段、任务、制品、触发人、耗时等）。

```bash
# 查询指定 run
codearts-cli pipeline status <pipeline_id> <pipeline_run_id> --project-id <project_id>

# 省略 pipeline_run_id，返回最近一次运行
codearts-cli pipeline status <pipeline_id> --project-id <project_id>
```

**API 参考**: [ShowPipelineRunDetail](https://support.huaweicloud.com/api-pipeline/ShowPipelineRunDetail.html)

| 参数 | 说明 |
| --- | --- |
| `<pipeline_id>` | 流水线 ID |
| `[pipeline_run_id]` | 可选；省略时返回最近一次运行 |
| `--project-id`（必填） | 华为云项目 UUID |
| `--dry-run` | 预览请求 |

**返回值**：`status`、`start_time`、`end_time`、`run_number`、`stages`（各阶段 step 状态）、`sources`、`artifacts` 等。

## 典型工作流

```bash
# 1. 列出可用流水线
codearts-cli pipeline list --project-id <proj>
#  → 得到 pipeline_id

# 2. 触发
RUN_ID=$(codearts-cli pipeline run <pid> --project-id <proj> 2>/dev/null | jq -r '.pipeline_run_id')

# 3. 查询运行状态
codearts-cli pipeline status <pid> $RUN_ID --project-id <proj>

# 4. 需要时停止
codearts-cli pipeline stop <pid> $RUN_ID --project-id <proj>
```

## 注意事项

- `--project-id` 对 pipeline 命令是**必填**的，不从 config.json 兜底——防止误操作到错误项目的流水线。
- 不传 `--sources` / `--variables` / `--body` 时，API 使用流水线存储的默认配置。
- `--sources` 和 `--variables` 是 JSON 数组格式，可用 `--sources-file` 从文件加载。
