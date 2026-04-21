# codearts-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)
[![npm version](https://img.shields.io/npm/v/@autelrobotics/codearts-cli.svg)](https://www.npmjs.com/package/@autelrobotics/codearts-cli)

华为云 [CodeArts](https://www.huaweicloud.com/product/codearts.html) 命令行工具，为人类和 AI Agent 而建。覆盖流水线、工作项管理、代码托管、编译构建四大模块共 19 个接口，配套 5 个 AI Agent [Skills](./skills/)。

[安装](#安装) · [AI Agent Skills](#agent-skills) · [配置](#配置) · [命令速查](#命令速查) · [高级用法](#高级用法) · [测试](#测试) · [架构](#项目结构) · [贡献](#贡献)

## 为什么用 codearts-cli？

- **Agent-Native 设计** — 4 个结构化 [Skills](./skills/)，兼容 Claude Code / Cursor / Codex / Gemini CLI 等主流 AI 工具
- **轻量零依赖** — 不引入 huaweicloud-sdk-go-v3（几十 MB），自研 AK/SK 签名（SDK-HMAC-SHA256），单一二进制 ~3 MB
- **四模块十七接口** — 流水线、工作项、代码托管、编译构建，一条命令触发 CI/CD、管理 Bug、创建 MR、跑编译
- **Debug 友好** — 所有命令支持 `--dry-run`，预览 method / path / body 不发请求
- **安全可控** — AK/SK 存储 `0600` 权限，`config show` 自动脱敏，CI 场景用 `--sk-stdin` 防泄露
- **开源即用** — MIT 协议，`npm install` 一行安装

## 功能概览

| 模块       | 命令                 | API                                | 说明                     |
| ---------- | -------------------- | ---------------------------------- | ------------------------ |
| 🚀 流水线  | `pipeline list`      | ListPipelines                      | 查询流水线列表           |
| 🚀 流水线  | `pipeline run`       | RunPipeline                        | 触发流水线               |
| 🚀 流水线  | `pipeline stop`      | StopPipelineRun                    | 停止流水线实例           |
| 🚀 流水线  | `pipeline status`    | ShowPipelineRunDetail              | 查询流水线运行详情       |
| 📋 工作项  | `issue list`         | ListIpdProjectIssues               | 查询工作项列表           |
| 📋 工作项  | `issue show`         | ShowIssueDetail                    | 查询工作项详情           |
| 📋 工作项  | `issue create`       | CreateIpdProjectIssue              | 创建工作项               |
| 📋 工作项  | `issue batch-update` | BatchUpdateIpdIssues               | 批量更新工作项           |
| 📋 工作项  | `issue relations`    | ListE2EGraphsOpenAPI               | 查询工作项关联（E2E 图） |
| 📋 工作项  | `issue members`      | ListProjectUsers                   | 查询项目成员             |
| 📋 工作项  | `issue statuses`     | ListIssueStatues                   | 查询工作项状态定义       |
| 🔀 代码托管 | `repo list`          | ShowAllRepositoryByTwoProjectId    | 查询仓库列表             |
| 🔀 代码托管 | `repo mr create`     | CreateMergeRequest                 | 创建合并请求             |
| 🔀 代码托管 | `repo mr comment`    | CreateMergeRequestDiscussion       | 创建 MR 检视意见         |
| 🔀 代码托管 | `repo member list`   | ListMembers                        | 查询仓库成员列表         |
| 🛠️ 编译构建 | `build list`         | ListProjectJobs                    | 查询项目构建任务列表     |
| 🛠️ 编译构建 | `build run`          | ExecuteJob                         | 触发构建                 |
| 🛠️ 编译构建 | `build stop`         | StopTheJob                         | 停止运行中的构建         |
| 🛠️ 编译构建 | `build status`       | ShowJobStepStatus                  | 查询构建任务步骤状态     |

## 安装

### 前置依赖

- Node.js (`npm`/`npx`)
- Go `v1.23`+（仅从源码构建时需要）

### 快速开始（人类用户）

> **AI Agent 提示：** 如果你是 AI Agent 正在帮用户安装，直接跳到 [快速开始（AI Agent）](#快速开始ai-agent)。

#### 安装

选择**一种**方式：

**方式 1 — npm 安装（推荐）：**

```bash
# 安装 CLI
npm install -g @autelrobotics/codearts-cli

# 安装 AI Agent Skills（可选但推荐）
npx skills add Lzhtommy/codearts-cli -y -g
```

**方式 2 — 从源码构建：**

需要 Go `v1.23`+。

```bash
git clone https://github.com/Lzhtommy/codearts-cli.git
cd codearts-cli
make install PREFIX=$HOME/.local

# 安装 AI Agent Skills（可选）
npx skills add Lzhtommy/codearts-cli -y -g
```

#### 配置 & 使用

```bash
# 1. 配置凭证（交互式，SK 不回显）
codearts-cli config init

# 2. 查看配置
codearts-cli config show

# 3. 开始使用
codearts-cli pipeline run <pipeline_id> --project-id <project_id>
codearts-cli issue list --issue-type Bug
```

### 快速开始（AI Agent）

> 以下步骤面向 AI Agent。部分步骤需要用户在终端中完成交互。

**Step 1 — 安装**

```bash
npm install -g @autelrobotics/codearts-cli
npx skills add Lzhtommy/codearts-cli -y -g
```

**Step 2 — 配置凭证**

> 交互式：用户需要在终端输入 AK、SK（不回显）、Project ID、Gateway URL、User ID。

```bash
codearts-cli config init
```

> 非交互（CI）：

```bash
echo "$HW_SK" | codearts-cli config init --ak "$HW_AK" --sk-stdin --project-id <uuid> --user-id <uuid> --yes
```

**Step 3 — 验证**

```bash
codearts-cli config show
codearts-cli issue list --issue-type Bug --dry-run
```

## Agent Skills

| Skill                | 说明                                                                              |
| -------------------- | --------------------------------------------------------------------------------- |
| `codearts-shared`    | 配置初始化、凭证管理、通用标志、端点解析、错误处理（被其它 skill 自动引用）       |
| `codearts-pipeline`  | 流水线列表 / 启动 / 停止 / 运行详情                                              |
| `codearts-issue`     | 工作项查询 / 详情 / 创建 / 批量更新 / 关联追溯 / 项目成员 / 状态定义             |
| `codearts-repo`      | 仓库列表 / MR 创建 / MR 检视意见 / 仓库成员                                      |
| `codearts-build`     | 构建任务列表 / 触发构建 / 停止构建 / 步骤状态                                    |

```bash
# 安装全部 skills
npx skills add Lzhtommy/codearts-cli -y -g

# 仅安装特定 skill
npx skills add Lzhtommy/codearts-cli -s codearts-pipeline -y -g
```

## 配置

### `config init`

交互式初始化，依次输入 5 项：

| # | 字段       | 必填 | 说明                                              |
| - | ---------- | ---- | ------------------------------------------------- |
| 1 | AK         | 是   | IAM Access Key ID                                 |
| 2 | SK         | 是   | IAM Secret Access Key（输入时不回显）             |
| 3 | Project ID | 默认 | CodeArts 项目 UUID（工作项接口必须，流水线/仓库可显式传入） |
| 4 | Gateway    | 默认 | 全局网关 URL，默认 `http://10.250.63.100:8099`    |
| 5 | User ID    | 可选 | IAM user_id（32 位 UUID），`issue create` 默认 assignee |

### `config set <key> <value>`

修改单个字段，无需走完整 init 流程：

```bash
codearts-cli config set userId <uuid>
codearts-cli config set gateway http://10.250.63.100:8099
```

可用 key：`ak` / `sk` / `projectId` / `gateway` / `userId`

### `config show` / `config path`

```bash
codearts-cli config show    # 打印配置（AK 脱敏、SK 全掩）
codearts-cli config path    # 打印配置文件绝对路径
```

配置文件：`~/.codearts-cli/config.json`（权限 `0600`）。

## 命令速查

### 流水线

#### `pipeline list` — 查询流水线列表

```bash
# 列出项目下所有流水线
codearts-cli pipeline list --project-id <proj>

# 按名称过滤 + 分页
codearts-cli pipeline list --project-id <proj> --name "deploy" --limit 20
```

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--name` | 按名称模糊匹配 |
| `--status` | 按状态过滤（可重复） |
| `--creator-id` / `--executor-id` | 按创建人/执行人过滤（可重复） |
| `--offset` / `--limit` | 分页 |
| `--sort-key` / `--sort-dir` | 排序（asc / desc） |
| `--dry-run` | 预览请求 |

#### `pipeline run <pipeline_id>` — 启动流水线

```bash
# 使用默认参数
codearts-cli pipeline run <pid> --project-id <proj>

# 指定源分支
codearts-cli pipeline run <pid> --project-id <proj> \
  --sources '[{"type":"code","params":{"build_type":"branch","target_branch":"main"}}]'

# 注入变量
codearts-cli pipeline run <pid> --project-id <proj> \
  --variables '[{"name":"ENV","value":"staging"}]'
```

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 华为云项目 UUID |
| `--sources` / `--sources-file` | 源覆盖 JSON 数组 |
| `--variables` / `--variables-file` | 变量 JSON 数组 |
| `--body` / `--body-file` | 完整请求体 |
| `--description` / `--choose-job` / `--choose-stage` | 备注 / 指定 job / stage |
| `--dry-run` | 预览请求 |

#### `pipeline stop <pipeline_id> <pipeline_run_id>` — 停止流水线

```bash
codearts-cli pipeline stop <pid> <run_id> --project-id <proj>
```

#### `pipeline status <pipeline_id> [pipeline_run_id]` — 查询运行详情

```bash
# 查指定 run
codearts-cli pipeline status <pid> <run_id> --project-id <proj>

# 省略 run_id，返回最近一次运行
codearts-cli pipeline status <pid> --project-id <proj>
```

返回 `status` / `start_time` / `end_time` / `run_number` / `stages` / `sources` / `artifacts` 等字段。

### 工作项管理

工作项命令统一使用 `config.json` 的 `projectId`，不支持 `--project-id`。若未配置，先执行 `codearts-cli config set projectId <uuid>`。

#### `issue list` — 查询列表

```bash
# 最简
codearts-cli issue list --issue-type Bug

# 查"我名下的 Bug"（assignee = config 中的 userId）
codearts-cli issue list --issue-type Bug \
  --filter '[{"assignee":{"values":["<your_user_id>"],"operator":"||"}}]'

# 分页 + 排序
codearts-cli issue list --issue-type US,Task --page-no 1 --page-size 50 --sort-field created_date
```

**filter 参数格式**：

```json
[{"<field>": {"values": ["..."], "operator": "||"}}]
```

- `<field>`：`assignee` / `status` / `priority` / … （加 `descendants.` 前缀可对子工作项下钻，一般无需）
- `operator`：`||`（OR，默认）/ `!`（NOT）/ `=` / `<>` / `<` / `>`

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 逗号分隔：`RR/SF/IR/US/Task/Bug/Epic/FE/SR/AR` |
| `--filter` / `--filter-file` | 过滤条件 JSON（格式见上） |
| `--filter-mode` | `AND_OR`（默认）/ `OR_AND` |
| `--page-no` / `--page-size` | 分页 |
| `--sort-field` / `--sort-asc` | 排序 |

#### `issue show <issue_id>` — 查询详情

```bash
codearts-cli issue show <id> --issue-type US
```

#### `issue create` — 创建工作项

```bash
codearts-cli issue create --title "修复登录超时" --description "..." --category Bug
```

`--assignee` 省略时自动取 config 中的 `userId`。

#### `issue batch-update` — 批量更新

```bash
codearts-cli issue batch-update --id 111,222,333 --category Bug --attribute '{"priority":"中"}'
```

#### `issue relations <issue_id>` — 查询工作项关联（E2E 图）

```bash
codearts-cli issue relations 1251275102548402177 --category US
```

返回父/子工作项、关联的 commits / MR / 分支 / 测试用例 / 文档。`--is-src true|false` 用于跨项目查询。

#### `issue members` — 查询项目成员

```bash
codearts-cli issue members
```

返回 `config.projectId` 下的所有成员，每条包含 `user_id`（32 位 UUID，可用作 `issue create --assignee`）、`user_name`、`nick_name`、`role_name`。

#### `issue statuses <category_id>` — 查询工作项状态定义

```bash
codearts-cli issue statuses 10020
```

`<category_id>` 是 5 位数字工作项类型 ID（非 Bug/Task 字符串）。有效取值：`10001` / `10020` / `10021` / `10022` / `10023` / `10027` / `10028` / `10029` / `10033` / `10065`。返回 `result` 数组，每条含 `name` 和 `belonging`（`START` / `IN_PROGRESS` / `END`）。

### 代码托管

> `<repository_id>` 必须是**正整数**（数字仓库 ID），不是 UUID。可通过 `repo list` 获取。

#### `repo list` — 查询仓库列表

```bash
# 列出项目下所有仓库（返回每个仓库的 repository_id）
codearts-cli repo list --project-id <proj>

# 按名称搜索
codearts-cli repo list --project-id <proj> --search "backend"
```

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 项目 UUID（可从 `git remote -v` 提取） |
| `--search` | 按仓库名或创建人搜索 |
| `--page-index` / `--page-size` | 分页（默认 20 条/页） |
| `--dry-run` | 预览请求 |

#### `repo mr create <repo_id>` — 创建合并请求

```bash
codearts-cli repo mr create 8147520 \
  --title "feat: x" --source feat/x --target main \
  --reviewers "uid-a,uid-b" --squash --force-remove-source
```

| Flag | 说明 |
| --- | --- |
| `--title` / `--source` / `--target`（必填） | 标题 / 源分支 / 目标分支 |
| `--reviewers` / `--assignees` | 逗号分隔 user_id |
| `--work-item` | 关联工作项 ID（可重复） |
| `--squash` / `--force-remove-source` | 合并选项 |
| `--body-json` / `--body-file` | 完整 JSON body |

#### `repo mr comment <repo_id> <mr_iid>` — 检视意见

```bash
codearts-cli repo mr comment 8147520 15 --body "LGTM" --severity suggestion
```

#### `repo member list <repo_id>` — 查询仓库成员

```bash
# 所有成员
codearts-cli repo member list 8147520

# 搜索 + 分页
codearts-cli repo member list 8147520 --search "zhang" --offset 0 --limit 50

# 按权限点过滤（如：有 code push 权限的成员）
codearts-cli repo member list 8147520 --permission code --action push
```

| Flag | 说明 |
| --- | --- |
| `--search` | 在 user_name / nick_name / tenant_name 上模糊匹配 |
| `--offset` / `--limit` | 分页（默认 offset=0，limit=20，limit 上限 100） |
| `--permission` | `repository` / `code` / `member` / `branch` / `tag` / `mr` / `label` |
| `--action` | 动作（需配合 `--permission`；取值随权限点不同，详见 skill 文档） |

### 编译构建

> `build list` 需要 `--project-id`（项目 UUID）。`build run` / `build stop` 使用 `<job_id>`（构建任务 ID，从 `build list` 返回值的 `id` 字段取）。

#### `build list` — 查询构建任务

```bash
# 列出项目下所有构建任务
codearts-cli build list --project-id <proj>

# 按名称 / 创建人模糊搜索 + 分页
codearts-cli build list --project-id <proj> --search "backend" \
  --page-index 0 --page-size 50

# 按最近一次构建状态过滤
codearts-cli build list --project-id <proj> --build-status red
```

| Flag | 说明 |
| --- | --- |
| `--project-id`（必填） | 项目 UUID |
| `--page-index` / `--page-size` | 分页（默认 0 / 10，page-size 上限 100） |
| `--search` | 任务名 / 创建人模糊匹配 |
| `--sort-field` / `--sort-order` | 排序 |
| `--creator-id` | 按创建人 user_id 过滤 |
| `--build-status` | `red` / `blue` / `timeout` / `aborted` / `building` / `none` |
| `--by-group` / `--group-path-id` | 分组查看 |

#### `build run <job_id>` — 触发构建

```bash
# 用任务默认参数触发
codearts-cli build run 48c66c6002964721be537cdc6ce0297b

# 覆盖分支 + 带参数 + 指定代码源
codearts-cli build run <job_id> \
  --branch main --build-type branch --scm-type codehub --repo-id 8147520 \
  --param "ENV=staging" --param "VERSION=1.2.0"
```

| Flag | 说明 |
| --- | --- |
| `--param KEY=VAL` | 构建参数（可重复或逗号分隔） |
| `--branch` / `--build-tag` / `--commit-id` | scm 源覆盖 |
| `--build-type` | `branch` / `tag` / `commitId` |
| `--scm-type` / `--repo-id` / `--repo-name` | 代码源：`default` / `codehub` |
| `--body-json` / `--body-file` | 完整 JSON body |

**返回值**包含 `daily_build_number`（后续 `build stop` 需要的 build_no）和 `actual_build_number`。

#### `build stop <job_id> <build_no>` — 停止构建

```bash
codearts-cli build stop 48c66c6002964721be537cdc6ce0297b 105
```

`<build_no>` 是**单次构建序号**（从 1 递增），来自 `build run` 返回值或构建历史面板，**不要**和 `job_id` 混淆。

#### `build status <job_id> [build_no]` — 查询构建状态

```bash
# 省略 build_no 时 API 默认查 build_no=1
codearts-cli build status <job_id>

# 查指定构建编号
codearts-cli build status <job_id> 42
```

返回 `result.workflow.status`（`completed`/`runnable`/`pending`）、`result.workflow.abort_status`（`aborted`/`timeout`/空）、`status`（`success`/`fail`）等字段。

## 高级用法

### Dry Run

所有命令支持 `--dry-run`，打印 method / path / query / body 但不发请求：

```bash
codearts-cli pipeline run <pid> --project-id <proj> --dry-run
codearts-cli issue create --title "x" --description "x" --category Bug --dry-run
```

### 网关

所有服务（流水线 / 工作项 / 代码托管 / 编译构建）统一走 `config.json` 中的 `gateway` 字段，默认 `http://10.250.63.100:8099`。切换网关：

```bash
codearts-cli config set gateway http://<your-gateway>:<port>
```

### project-id 作用域

| 模块       | `--project-id` | config.json `projectId` |
| ---------- | -------------- | ----------------------- |
| 流水线     | **必填**       | 不读                    |
| 工作项管理 | 不支持         | **必读**                |
| 代码托管   | `repo list` 必填；`mr` / `member` 命令走 `repo_id` | `repo list` 不读 |
| 编译构建   | `build list` 必填；`run` / `stop` 走 `job_id` | 不读 |

> **提示**：在 CodeArts Repo 克隆的仓库目录下，可从 `git remote -v` 自动提取 project-id：
> ```bash
> PROJECT_ID=$(git remote -v | grep codehub | head -1 | grep -oE '[a-f0-9]{32}' | head -1)
> ```

## 测试

单元测试集中在 `tests/` 目录下，按模块组织：

```bash
# 运行全部测试
go test ./tests/...

# 运行单个模块
go test ./tests/client/    # 签名 + HTTP 客户端
go test ./tests/core/      # 配置加载/保存/校验
go test ./tests/output/    # JSON 输出格式
go test ./tests/cmd/       # CLI 工具函数

# 带详细输出
go test ./tests/... -v

# 跑完整项目（含 tests/）
go test ./...
```

| 测试模块         | 文件                          | 用例数 | 覆盖点                                                            |
| ---------------- | ----------------------------- | ------ | ------------------------------------------------------------------ |
| `tests/client/`  | `signer_test.go`              | 11     | HashHex / HmacHex / CanonicalURI / CanonicalQuery / CanonicalHeaders / Sign |
| `tests/client/`  | `client_test.go`              | 7      | Do 成功 / 400 / 401 hint / 500 raw / 空响应 / POST body / 缺凭证  |
| `tests/core/`    | `config_test.go`              | 10     | Validate / Redacted / MaskLeft / Save+Load 往返 / 默认值 / 兼容性  |
| `tests/output/`  | `output_test.go`              | 5      | PrintJSON / Successf / Errorf / DryRunf                            |
| `tests/cmd/`     | `helpers_test.go`             | 14     | ParseRepoID / ExtractStringFromResp / FirstNonEmpty                |
| **合计**         |                               | **47** |                                                                    |

HTTP 测试使用 `httptest.Server` 模拟 API，文件 I/O 测试使用 `t.TempDir()` + 临时 `$HOME`，不依赖外部服务。

## 项目结构

```
codearts-cli/
├── main.go                           # 入口
├── Makefile                          # build / install / dist / npm-link
├── package.json                      # npm 包描述
├── .goreleaser.yml                   # 跨平台编译配置
├── scripts/
│   ├── install.js                    # npm postinstall：下载预编译二进制
│   └── run.js                        # npm bin 入口：exec Go 二进制
├── skills/
│   ├── codearts-shared/SKILL.md      # AI Skill: 配置 / 认证 / 通用
│   ├── codearts-pipeline/SKILL.md    # AI Skill: 流水线
│   ├── codearts-issue/SKILL.md       # AI Skill: 工作项
│   ├── codearts-repo/SKILL.md        # AI Skill: 代码托管
│   └── codearts-build/SKILL.md       # AI Skill: 编译构建
├── tests/
│   ├── client/                       # 签名 + HTTP 客户端测试
│   ├── core/                         # 配置测试
│   ├── output/                       # 输出格式测试
│   └── cmd/                          # CLI 工具函数测试
├── cmd/
│   ├── root.go                       # 根命令
│   ├── config.go                     # config init / show / path / set
│   ├── pipeline.go                   # pipeline list / run / stop / status
│   ├── issue.go                      # issue list / show / create / batch-update / relations / members / statuses
│   ├── repo.go                       # repo list / mr create / mr comment / member list
│   └── build.go                      # build list / run / stop / status
└── internal/
    ├── core/config.go                # 配置加载 / 保存（~/.codearts-cli/config.json）
    ├── client/
    │   ├── signer.go                 # SDK-HMAC-SHA256 AK/SK 签名
    │   ├── client.go                 # HTTP 封装 + 三服务端点推导
    │   ├── pipeline.go               # 流水线 API
    │   ├── projectman.go             # 工作项 API
    │   ├── repo.go                   # 代码托管 API
    │   └── build.go                  # 编译构建 API
    └── output/output.go              # JSON 输出 + 成功 / 错误消息
```

## 扩展新 API

1. 如果是新服务域，在 `internal/client/client.go` 加 `XxxEndpoint()` 方法
2. 在 `internal/client/*.go` 加接口方法，调用 `c.Do(ctx, method, endpoint, path, query, body, &out)`
3. 在 `cmd/` 加 cobra 子命令，参考 `--dry-run` 模式
4. 在 `cmd/root.go` 注册

已踩坑记录：
- 华为 APIGW 签名的 CanonicalURI **必须以 `/` 结尾**
- POST 即使无参数也必须发 `{}`（否则 `PARSE_REQUEST_DATA_EXCEPTION`）
- 401 `APIGW.0301` 会回显服务端 `canonical_request`，直接 diff 即可定位

## 贡献

欢迎社区贡献！如果发现 Bug 或有功能建议，请提交 [Issue](https://github.com/Lzhtommy/codearts-cli/issues) 或 [Pull Request](https://github.com/Lzhtommy/codearts-cli/pulls)。

重大变更建议先通过 Issue 讨论。

## License

本项目基于 **MIT License** 开源。运行时调用华为云 CodeArts 平台 API，使用前请遵守 [华为云服务协议](https://www.huaweicloud.com/declaration/sa.html) 和 [华为云隐私政策](https://www.huaweicloud.com/declaration/tsa.html)。
