---
name: codearts-issue
version: 0.1.1
description: "CodeArts 工作项管理（ProjectMan IPD）：查询工作项列表、查询详情、创建工作项、批量更新。当用户需要管理 Bug/Task/US/Epic 等工作项时使用。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli issue --help"
---

# codearts-issue (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../codearts-shared/SKILL.md`](../codearts-shared/SKILL.md) 了解配置与认证。**

CodeArts ProjectMan 工作项（IPD）管理。**所有命令均使用 `config.json` 中的 `projectId`**——若未配置，先执行 `codearts-cli config set projectId <uuid>`。

## issue_type 取值

不同项目类型支持的 issue_type 不同：

| 项目类型       | 支持的 issue_type                         |
| -------------- | ----------------------------------------- |
| 系统设备       | RR, SF, IR, SR, AR, Task, Bug             |
| 独立软件       | RR, SF, IR, US, Task, Bug                 |
| 云服务         | RR, Epic, FE, US, Task, Bug              |

## 命令

### issue list

查询项目工作项列表。`--issue-type` 必填。

```bash
# 最简
codearts-cli issue list --issue-type Bug

# 多类型 + 分页 + 过滤
codearts-cli issue list --issue-type US,Task \
  --page-no 1 --page-size 50 \
  --filter '[{"property":"status","condition":"include","value":["new"]}]'

# 排序
codearts-cli issue list --issue-type Bug --sort-field created_date --sort-asc
```

**API**: `POST /v1/ipdprojectservice/projects/{project_id}/issues/query?issue_type=...`

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 逗号分隔的类型列表 |
| `--filter` / `--filter-file` | JSON 数组过滤条件 |
| `--filter-mode` | `AND_OR`（默认）/ `OR_AND` |
| `--page-no` / `--page-size` | 分页（0 = API 默认） |
| `--sort-field` / `--sort-asc` | 排序字段与方向 |
| `--dry-run` | 预览请求 |

### issue show

查询单个工作项详情。

```bash
codearts-cli issue show <issue_id> --issue-type US
```

**API**: `GET /v1/ipdprojectservice/projects/{project_id}/issues/{issue_id}?issue_type=...`

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 工作项类型 |
| `--domain-id` | 可选 |
| `--dry-run` | 预览请求 |

### issue create

创建工作项。

```bash
# 最简（assignee 从 config userId 自动填充）
codearts-cli issue create \
  --title "修复登录超时" \
  --description "用户反馈在弱网环境下登录超时" \
  --category Bug

# 显式指定 assignee
codearts-cli issue create \
  --title "接入 CodeArts CLI" \
  --description "完成 AK/SK 与 RunPipeline 接入" \
  --category US \
  --assignee <user_id_32char>

# 完整字段用 body-file
codearts-cli issue create --body-file issue.json
```

**API**: `POST /v1/ipdprojectservice/projects/{project_id}/issues`

| Flag | 说明 |
| --- | --- |
| `--title`（必填*） | 标题，最长 256 字符 |
| `--description`（必填*） | 描述，最长 500000 字符 |
| `--category`（必填*） | 类型：RR/SF/IR/SR/AR/Task/Bug/US/Epic/FE |
| `--assignee` | user_id UUID；省略时从 config `userId` 取 |
| `--status` / `--priority` | 可选 |
| `--body` / `--body-file` | 完整 JSON（覆盖上面所有 flag） |
| `--dry-run` | 预览请求 |

*使用 `--body` / `--body-file` 时不需要这些 flag。

### issue batch-update

批量更新工作项。

```bash
# 更新多个 issue 的 priority
codearts-cli issue batch-update \
  --id 111,222 --id 333 \
  --category Bug \
  --attribute '{"priority":"中"}'
```

**API**: `PUT /v1/ipdprojectservice/projects/{project_id}/issues/batch`

| Flag | 说明 |
| --- | --- |
| `--id`（必填） | issue ID，可重复或逗号分隔 |
| `--category`（必填*） | 目标工作项类型 |
| `--attribute` / `--attribute-file` | JSON 对象：要更新的属性 |
| `--dry-run` | 预览请求 |

*category 也可在 `--attribute` JSON 中提供。

## 常见错误

- **PM.02177003 非目标项目成员**：assignee 的 user_id 不是项目成员。注意不要把 tenant_id 当 user_id（格式相同但含义不同）。
- **issue_type 不支持**：检查项目类型（系统设备 / 独立软件 / 云服务）是否支持该 issue_type。
