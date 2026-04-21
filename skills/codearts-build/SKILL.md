---
name: codearts-build
version: 0.1.0
description: "CodeArts 编译构建（CodeCI）：查询构建任务列表（ListProjectJobs）、执行构建（ExecuteJob）、停止构建（StopTheJob）。当用户需要查看构建任务、触发构建或停止运行中的构建时使用。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli build --help"
---

# codearts-build (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../codearts-shared/SKILL.md`](../codearts-shared/SKILL.md) 了解配置与认证。**

CodeArts 编译构建（CodeCI）模块。`build list` 必须显式传 `--project-id`（不从 config 兜底）；`build run` / `build stop` 只需要 `job_id` 和 `build_no`。

## 名词

| 概念 | 说明 |
| --- | --- |
| `project_id` | 32 位 CodeArts 项目 UUID。可从 `git remote -v` 提取或 CodeArts 控制台 URL 中拿到 |
| `job_id` | 32 位构建任务 ID（`build list` 返回值中 `id` 字段） |
| `build_no` | 单次构建实例编号（从 1 开始递增）。`build run` 返回体里的 `daily_build_number` / `actual_build_number` |

## 命令

### build list

查询项目下的构建任务列表。

```bash
# 列出所有任务
codearts-cli build list --project-id <proj>

# 按名称 / 创建人模糊搜索 + 分页
codearts-cli build list --project-id <proj> --search "backend" \
  --page-index 0 --page-size 50

# 按最近一次构建状态过滤
codearts-cli build list --project-id <proj> --build-status red

# 按分组
codearts-cli build list --project-id <proj> --by-group --group-path-id <gid>
```

**API**: `GET /v1/job/{project_id}/list`

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 项目 UUID |
| `--page-index` | 起始页（0-based，默认 0） |
| `--page-size` | 每页条数（1-100，默认 10） |
| `--search` | 在任务名 / 创建人上模糊匹配 |
| `--sort-field` / `--sort-order` | 排序字段 / 方向（asc\|desc） |
| `--creator-id` | 按创建人 user_id 过滤 |
| `--build-status` | `red` / `blue` / `timeout` / `aborted` / `building` / `none` |
| `--by-group` / `--group-path-id` | 分组查看 |
| `--dry-run` | 预览请求 |

**返回值**：`result.job_list` 数组，每条包含：
- `id`（**用于 `build run` / `build stop`**）、`job_name`、`job_creator`、`user_name`
- `last_build_time`、`last_build_status`（`red`/`blue`/`timeout`/`aborted`）、`build_number`、`is_finished`
- 权限位：`is_modify`、`is_delete`、`is_execute`、`is_copy`、`is_view`
- 关联源：`scm_type`（codehub\|repo\|github）、`repo_id`、`commit_id`

### build run

触发一次构建。

```bash
# 最简：用任务保存的默认参数触发
codearts-cli build run <job_id>

# 覆盖源分支 / 构建类型
codearts-cli build run <job_id> \
  --branch main --build-type branch \
  --scm-type codehub --repo-id 8147520

# 传构建参数（可重复或逗号分隔）
codearts-cli build run <job_id> \
  --param "ENV=staging" --param "VERSION=1.2.0"

# 用 tag 或 commit 触发
codearts-cli build run <job_id> --build-type tag --build-tag v1.2.0
codearts-cli build run <job_id> --build-type commitId --commit-id 7a9f...

# 完整 body
codearts-cli build run <job_id> --body-file build.json
```

**API**: `POST /v1/job/execute`（`job_id` 写入 body）

| Flag | 说明 |
| --- | --- |
| `--param KEY=VAL` | 构建参数（可重复，逗号分隔也可） |
| `--branch` | scm.branch |
| `--build-tag` | scm.build_tag |
| `--commit-id` | scm.build_commit_id |
| `--build-type` | `branch` / `tag` / `commitId` |
| `--repo-id` / `--repo-name` | scm.repo_id / scm.repo_name |
| `--scm-type` | `default` / `codehub` |
| `--scm-url` / `--scm-web-url` | scm.url / scm.web_url |
| `--body-json` / `--body-file` | 完整 JSON body（会保留 job_id 覆盖） |
| `--dry-run` | 预览请求 |

**返回值**：`octopus_job_name`、`actual_build_number`、`daily_build_number`。`daily_build_number` 就是 `build stop` 需要的 `<build_no>`。

### build stop

停止一次运行中的构建。

```bash
codearts-cli build stop <job_id> <build_no>
```

**API**: `POST /v1/job/{job_id}/stop`，body `{"build_no": <int>}`。

| 参数 | 说明 |
| --- | --- |
| `<job_id>` | 32 位构建任务 ID |
| `<build_no>` | 构建编号（正整数 >= 1，来自 `build run` 返回） |
| `--dry-run` | 预览请求 |

**返回值**：`{"status": "success"}`。

## 典型工作流

```bash
# 1. 查任务 id
codearts-cli build list --project-id <proj> --search "backend"
# → id: "48c66c6002964721be537cdc6ce0297b"

# 2. 触发构建，拿到 build_no
JOB_ID=48c66c6002964721be537cdc6ce0297b
BUILD_NO=$(codearts-cli build run $JOB_ID --branch main --build-type branch 2>/dev/null \
  | jq -r '.daily_build_number')

# 3.（必要时）停止
codearts-cli build stop $JOB_ID $BUILD_NO
```

## 常见坑

- `build list` 的 `--project-id` 是 32 位 UUID，不是中文项目名，也不是数字 ID
- `build run` 不传任何 `--branch` / `--param` / `scm-*` 也可以 —— 此时用任务保存的默认触发，仅把 `job_id` 写入 body
- `build stop` 的 `<build_no>` 是**单次构建序号**（从 1 开始），不是 `job_id`，也不是 pipeline_run_id
- `build_status` 枚举：`red`=失败 / `blue`=成功 / `timeout` / `aborted` / `building`=运行中 / `none`=未构建
