---
name: codearts-repo
version: 0.1.2
description: "CodeArts 代码托管：查询仓库列表（ListRepositories）、创建合并请求（CreateMergeRequest）、创建 MR 检视意见（CreateMergeRequestDiscussion）、查询仓库成员（ListMembers）。当用户需要查看仓库、创建 MR、发代码评审意见或查询仓库成员时使用。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli repo --help"
---

# codearts-repo (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../codearts-shared/SKILL.md`](../codearts-shared/SKILL.md) 了解配置与认证。**

CodeArts 代码托管模块。`--project-id` 在所有 repo 命令中是**必填**的（不从 config 兜底）。

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

HTTPS 格式同理：

```
https://codehub-cn-south-1.devcloud.huaweicloud.com/759278abbfb14b098eeddc548741f38b/nest-app-agent.git
```

```bash
PROJECT_ID=$(git remote -v | grep codehub | head -1 | grep -oE '[a-f0-9]{32}' | head -1)
```

提取后直接传给 `--project-id $PROJECT_ID`。如果 `git remote -v` 中没有 `codehub` 的 URL，则需要用户手动提供。

## 命令

### repo list

查询项目下的仓库列表。`--project-id` 必填（可从 `git remote -v` 自动提取）。

```bash
# 列出所有仓库
codearts-cli repo list --project-id <project_uuid>

# 按名称搜索
codearts-cli repo list --project-id <project_uuid> --search "backend"

# 分页
codearts-cli repo list --project-id <project_uuid> --page-index 2 --page-size 10
```

**API**: `GET /v2/projects/{project_uuid}/repositories`

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 项目 UUID（从 `git remote -v` 提取或手动传入） |
| `--search` | 按仓库名或创建人名搜索 |
| `--page-index` | 页码（1-based，0 = 默认第 1 页） |
| `--page-size` | 每页条数（1-100，默认 20） |
| `--dry-run` | 预览请求 |

**返回值**：`result.repositories` 数组，每条包含 `repository_id`（**整数**，用于 MR 操作）、`repository_name`、`ssh_url`、`https_url`、`web_url`。

> **关键**：`repo list` 返回的 `repository_id` 就是 `repo mr create` / `repo mr comment` 需要的参数。

## 重要：repository_id 是整数

所有 repo 命令的 `<repository_id>` 必须是**正整数**（如 `8147520`），**不是** 32 位 UUID。

- UUID 格式（如 `759278abbfb14b098eeddc548741f38b`）是 **project_id**，不是 repo_id
- 获取 repo_id：运行 `codearts-cli repo list --project-id <proj>`，或 CodeArts Repo 控制台 → 仓库设置 → 仓库 ID（数字）

CLI 会**严格校验**：传入 UUID 会直接报错（不会静默截断）。

## 命令

### repo mr create

创建合并请求。

```bash
# 最简
codearts-cli repo mr create <repo_id> \
  --title "feat: 接入 codearts-cli" \
  --source feat/cli --target main

# 带评审人 / squash / 关联工作项
codearts-cli repo mr create <repo_id> \
  --title "feat: x" --source feat/x --target main \
  --reviewers "uid-a,uid-b" --assignees "uid-c" \
  --squash --squash-message "feat: squashed" \
  --force-remove-source \
  --work-item 1251275102548402177

# 完整 body
codearts-cli repo mr create <repo_id> --body-file mr.json
```

**API**: `POST /v4/repositories/{repository_id}/merge-requests`

| Flag | 说明 |
| --- | --- |
| `--title`（必填*） | MR 标题 |
| `--source`（必填*） | 源分支 |
| `--target`（必填*） | 目标分支 |
| `--description` | MR 描述 |
| `--reviewers` | 评审人 user_id 逗号分隔 |
| `--assignees` | 指派人 user_id 逗号分隔 |
| `--approval-reviewers` | 审批评审人 |
| `--approval-approvers` | 审批人 |
| `--work-item` | 关联工作项 ID（可重复或逗号分隔） |
| `--milestone-id` | 里程碑 ID |
| `--squash` / `--squash-message` | 合并时 squash commits |
| `--force-remove-source` | 合并后自动删除源分支 |
| `--only-assignee-merge` | 仅允许指派人合入 |
| `--target-repo-id` | 跨仓库 MR 的目标仓库 ID |
| `--body-json` / `--body-file` | 完整 JSON body |
| `--dry-run` | 预览请求 |

*使用 `--body-json` / `--body-file` 时不需要这些 flag。

**返回值**包含 `iid`（MR 编号，用于后续 `repo mr comment`）和 `web_url`（控制台链接）。

### repo mr comment

给合并请求发检视意见。

```bash
# 简单评论
codearts-cli repo mr comment <repo_id> <mr_iid> --body "LGTM"

# 带严重级别
codearts-cli repo mr comment <repo_id> <mr_iid> \
  --body "参数未校验" --severity major

# 行级评论（需要 position 结构，用文件）
codearts-cli repo mr comment <repo_id> <mr_iid> --body-file review.json
```

**API**: `POST /v4/repositories/{repository_id}/merge-requests/{merge_request_iid}/discussions`

| Flag | 说明 |
| --- | --- |
| `--body`（必填*） | 评论内容 |
| `--severity` | `suggestion` / `minor` / `major` / `fatal` |
| `--assignee-id` | 指派人 |
| `--review-categories` | 评审分类 |
| `--review-modules` | 评审模块 |
| `--proposer-id` | 评审发起人 |
| `--line-types` | 行类型（行级评论场景） |
| `--body-json` / `--body-file` | 完整 JSON body（行级评论需要 position） |
| `--dry-run` | 预览请求 |

*使用 `--body-json` / `--body-file` 时不需要此 flag。

### repo member list

查询仓库成员列表。`<repository_id>` 必须是**正整数**（不是 UUID）。

```bash
# 所有成员
codearts-cli repo member list <repo_id>

# 按用户名 / 昵称 / 租户名模糊搜索
codearts-cli repo member list <repo_id> --search "zhang"

# 分页（offset 从 0 开始，limit 1-100，默认 20）
codearts-cli repo member list <repo_id> --offset 20 --limit 50

# 按权限点 + 动作过滤（如：有 code push 权限的成员）
codearts-cli repo member list <repo_id> --permission code --action push
```

**API**: `GET /v4/repositories/{repository_id}/members`

| Flag | 说明 |
| --- | --- |
| `--search` | 在 user_name / user_nick_name / tenant_name 上模糊匹配 |
| `--offset` | 分页偏移（0-based，默认 0） |
| `--limit` | 每页条数（1-100，默认 20） |
| `--permission` | 权限点：`repository` / `code` / `member` / `branch` / `tag` / `mr` / `label` |
| `--action` | 动作（需配合 `--permission`，取值随权限点不同） |
| `--dry-run` | 预览请求 |

**`--action` 取值**（按 `--permission` 分）：

| permission | action |
| --- | --- |
| `repository` | `create` / `fork` / `delete` / `setting` |
| `code` | `push` / `download` |
| `member` | `create` / `update` / `delete` |
| `branch` | `create` / `delete` |
| `tag` | `create` / `delete` |
| `mr` | `create` / `update` / `comment` / `review` / `approve` / `merge` / `close` / `reopen` |
| `label` | `create` / `update` / `delete` |

**返回值**：成员 DTO 数组，字段包括 `user_id`、`user_name`、`user_nick_name`、`tenant_name`、`repository_role_name`、`service_license_status`（0=停用 / 1=正常）等。

## 典型工作流

```bash
# 1. 查询仓库列表，获取 repository_id
codearts-cli repo list --project-id <proj>
#  → repository_id: 8147520, repository_name: "nest-app-agent"

# 2. 创建 MR
MR_IID=$(codearts-cli repo mr create 8147520 \
  --title "feat: x" --source feat/x --target main 2>/dev/null \
  | jq -r '.iid')

# 3. 发检视意见
codearts-cli repo mr comment 8147520 $MR_IID --body "请补单测" --severity minor
```
