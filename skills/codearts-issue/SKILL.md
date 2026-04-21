---
name: codearts-issue
version: 0.1.3
description: "CodeArts 工作项管理（ProjectMan IPD）：查询工作项列表、查询详情、创建工作项、批量更新、查询工作项关联、查询项目成员、查询工作项状态。当用户需要管理 Bug/Task/US/Epic 等工作项、查看项目成员或查询某工作项类型的状态定义时使用。"
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

# 查"我名下的 Bug"（assignee = config 中的 userId）
codearts-cli issue list --issue-type Bug \
  --filter "[{\"assignee\":{\"values\":[\"$(codearts-cli config show | jq -r .userId)\"],\"operator\":\"||\"}}]"

# 多类型 + 分页 + 排序
codearts-cli issue list --issue-type US,Task \
  --page-no 1 --page-size 50 \
  --sort-field created_date --sort-asc
```

**API**: `POST /v1/ipdprojectservice/projects/{project_id}/issues/query?issue_type=...`

### filter 参数结构

```json
[ { "<字段名>": { "values": ["..."], "operator": "||" } } ]
```

数组的每个元素是一个以字段名为 key 的 map；值是 `ConditionVO`。

| 字段名示例 | 含义 |
| --- | --- |
| `assignee` | 处理人 user_id |
| `status` | 状态 |
| `priority` | 优先级（中 / 高 / 低） |
| `descendants.<field>` | 同名字段的树形下钻版；一般场景用裸字段即可 |

`operator` 取值：`||`（OR，默认）、`!`（NOT）、`=`（等于单值）、`<>` / `<` / `>`（日期/数字范围）。

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 逗号分隔的类型列表 |
| `--filter` / `--filter-file` | JSON 数组过滤条件（格式见上） |
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
| `--status` | 可选。合法值：`Committed` / `Analyse` / `ToBeConfirmed` / `Plan` / `Doing` / `Delivered` / `Checking` |
| `--priority` | 可选。合法值通常为 `中` / `高` / `低`（项目自定义） |
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

### issue relations

查询工作项的端到端追溯关系（E2E 图）——父/子工作项、关联提交 / MR / 分支 / 测试用例 / 测试计划 / 文档。

```bash
# 查询一个 US 的追溯图
codearts-cli issue relations <issue_id> --category US

# 跨项目（上游 / 下游）
codearts-cli issue relations <issue_id> --category Bug --is-src true
```

**API**: `GET /v1/ipdprojectservice/projects/{project_id}/e2e/graphs?issue_id=&category=&is_src=`

| Flag | 说明 |
| --- | --- |
| `<issue_id>`（位置参数） | 18–19 位数字 ID（不是控制台看到的 `number` 短号，是 API 返回的 `id` 字段） |
| `--category`（必填） | RR/SF/IR/SR/AR/Task/Bug/US/Epic/FE |
| `--is-src` | `true` / `false`，跨项目查询方向；省略则按 API 默认 |
| `--dry-run` | 预览请求 |

**返回值**：`id`、`project_id`、`domain_id`、`category`、`number`、`status`（初始/分析/测试/开发/完成）、`title`，以及 `trace_list` 数组 —— 元素包含：
- `parent_issues` / `child_issues` — 父/子工作项
- `associate_workitems` — 关联的其它工作项
- `associate_commits` / `associate_branches` / `associate_mergerequest` — 关联的代码资产
- `associate_testcases` / `associate_testplans` — 关联的测试资产
- `associate_documents` — 关联的文档

### issue members

查询当前 `projectId` 下的所有项目成员。

```bash
codearts-cli issue members
```

**API**: `GET /v1/ipdprojectservice/projects/{project_id}/users`

| Flag | 说明 |
| --- | --- |
| `--dry-run` | 预览请求 |

**返回值**：`result` 数组，每条 `UserVO` 包含：
- `user_id`（**32 位 UUID**，用于 `issue create --assignee` 和 `issue list --filter` 的 assignee 字段）
- `user_num_id`（整数短 ID）
- `user_name`、`nick_name`、`domain_id`、`domain_name`（租户名）
- `role_id` / `role_name`（多个角色逗号分隔）

> **典型用法**：给 `issue create` 找 `--assignee` 时，先跑 `issue members | jq '.result[] | {user_id, user_name, nick_name}'` 拿到真实的 user_id —— 不要把 tenant_id 当 user_id 传（格式都是 32 位 UUID 但含义不同，会触发 `PM.02177003 非目标项目成员`）。

### issue statuses

查询某个工作项类型（category_id）在项目里配置的状态定义。

```bash
codearts-cli issue statuses <category_id>
```

**API**: `GET /v1/ipdprojectservice/projects/{project_id}/category/{category_id}/statuses`

| 参数 | 说明 |
| --- | --- |
| `<category_id>`（位置参数） | **5 位纯数字**工作项类型 ID，不是 RR/Bug/Task 字符串 |
| `--dry-run` | 预览请求 |

**有效 `category_id` 取值**（API 文档枚举）：`10001` / `10020` / `10021` / `10022` / `10023` / `10027` / `10028` / `10029` / `10033` / `10065`。

> **注意**：字符串分类名（Bug/Task/US/…）与 `category_id` 的映射是项目级配置，不同项目不同 —— 不要硬编码。在 CodeArts Req 控制台 **工作项类型** 设置页或其它接口的返回里查一次，本地记下来。

**返回值**：`result` 数组，每条含 `name`（状态名，如 "新建"/"开发中"/"已关闭"）与 `belonging`（生命周期分桶：`START` / `IN_PROGRESS` / `END`）。

## 常见错误

- **PM.02177003 非目标项目成员**：assignee 的 user_id 不是项目成员。注意不要把 tenant_id 当 user_id（格式相同但含义不同）。
- **issue_type 不支持**：检查项目类型（系统设备 / 独立软件 / 云服务）是否支持该 issue_type。
