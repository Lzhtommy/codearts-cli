# codearts-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://go.dev/)
[![npm version](https://img.shields.io/npm/v/@autelrobotics/codearts-cli.svg)](https://www.npmjs.com/package/@autelrobotics/codearts-cli)

华为云 [CodeArts](https://www.huaweicloud.com/product/codearts.html) 命令行工具，为人类和 AI Agent 而建。覆盖流水线、工作项管理、代码托管三大模块共 8 个接口，配套 4 个 AI Agent [Skills](./skills/)。

[安装](#安装) · [AI Agent Skills](#agent-skills) · [配置](#配置) · [命令速查](#命令速查) · [高级用法](#高级用法) · [架构](#项目结构) · [贡献](#贡献)

## 为什么用 codearts-cli？

- **Agent-Native 设计** — 4 个结构化 [Skills](./skills/)，兼容 Claude Code / Cursor / Codex / Gemini CLI 等主流 AI 工具
- **轻量零依赖** — 不引入 huaweicloud-sdk-go-v3（几十 MB），自研 AK/SK 签名（SDK-HMAC-SHA256），单一二进制 ~3 MB
- **三模块八接口** — 流水线、工作项、代码托管，一条命令触发 CI/CD、管理 Bug、创建 MR
- **Debug 友好** — 所有命令支持 `--dry-run`，预览 method / path / body 不发请求
- **安全可控** — AK/SK 存储 `0600` 权限，`config show` 自动脱敏，CI 场景用 `--sk-stdin` 防泄露
- **开源即用** — MIT 协议，`npm install` 一行安装

## 功能概览

| 模块       | 命令                 | API                          | 说明                     |
| ---------- | -------------------- | ---------------------------- | ------------------------ |
| 🚀 流水线  | `pipeline run`       | RunPipeline                  | 触发流水线               |
| 🚀 流水线  | `pipeline stop`      | StopPipelineRun              | 停止流水线实例           |
| 📋 工作项  | `issue list`         | ListIpdProjectIssues         | 查询工作项列表           |
| 📋 工作项  | `issue show`         | ShowIssueDetail              | 查询工作项详情           |
| 📋 工作项  | `issue create`       | CreateIpdProjectIssue        | 创建工作项               |
| 📋 工作项  | `issue batch-update` | BatchUpdateIpdIssues         | 批量更新工作项           |
| 🔀 代码托管 | `repo mr create`     | CreateMergeRequest           | 创建合并请求             |
| 🔀 代码托管 | `repo mr comment`    | CreateMergeRequestDiscussion | 创建 MR 检视意见         |

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

> 交互式：用户需要在终端输入 AK、SK（不回显）、Project ID、Region、User ID。

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
| `codearts-pipeline`  | 流水线启动 / 停止                                                                 |
| `codearts-issue`     | 工作项查询 / 详情 / 创建 / 批量更新                                              |
| `codearts-repo`      | MR 创建 / MR 检视意见                                                            |

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
| 3 | Project ID | 默认 | CodeArts 项目 UUID（仅工作项接口使用此默认值）    |
| 4 | Region     | 默认 | 如 `cn-south-1`                                   |
| 5 | User ID    | 可选 | IAM user_id（32 位 UUID），`issue create` 默认 assignee |

### `config set <key> <value>`

修改单个字段，无需走完整 init 流程：

```bash
codearts-cli config set userId <uuid>
codearts-cli config set region cn-north-4
```

可用 key：`ak` / `sk` / `projectId` / `region` / `userId`

### `config show` / `config path`

```bash
codearts-cli config show    # 打印配置（AK 脱敏、SK 全掩）
codearts-cli config path    # 打印配置文件绝对路径
```

配置文件：`~/.codearts-cli/config.json`（权限 `0600`）。

## 命令速查

### 流水线

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

### 工作项管理

`--project-id` 可选，省略时从 `config.json` 的 `projectId` 兜底。

#### `issue list` — 查询列表

```bash
codearts-cli issue list --issue-type US,Task --page-no 1 --page-size 50
```

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 逗号分隔：`RR/SF/IR/US/Task/Bug/Epic/FE/SR/AR` |
| `--filter` / `--filter-file` | 过滤条件 JSON |
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

### 代码托管

> `<repository_id>` 必须是**正整数**（数字仓库 ID），不是 UUID。

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

## 高级用法

### Dry Run

所有命令支持 `--dry-run`，打印 method / path / query / body 但不发请求：

```bash
codearts-cli pipeline run <pid> --project-id <proj> --dry-run
codearts-cli issue create --title "x" --description "x" --category Bug --dry-run
```

### 端点覆盖

各模块按 region 自动推导端点；需要时用环境变量覆盖（如抓包调试、私有化部署）：

| 模块       | 默认主机                                       | 环境变量                         |
| ---------- | ---------------------------------------------- | -------------------------------- |
| 流水线     | `cloudpipeline-ext.<region>.myhuaweicloud.com` | `CODEARTS_PIPELINE_ENDPOINT`     |
| 工作项管理 | `projectman-ext.<region>.myhuaweicloud.com`    | `CODEARTS_PROJECTMAN_ENDPOINT`   |
| 代码托管   | `codehub-ext.<region>.myhuaweicloud.com`       | `CODEARTS_REPO_ENDPOINT`         |

### project-id 作用域

| 模块       | `--project-id` | config.json `projectId` |
| ---------- | -------------- | ----------------------- |
| 流水线     | **必填**       | 不读                    |
| 工作项管理 | 可选覆盖       | 兜底                    |
| 代码托管   | 不涉及         | 不涉及（走 repo_id）   |

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
│   └── codearts-repo/SKILL.md        # AI Skill: 代码托管
├── cmd/
│   ├── root.go                       # 根命令
│   ├── config.go                     # config init / show / path / set
│   ├── pipeline.go                   # pipeline run / stop
│   ├── issue.go                      # issue list / show / create / batch-update
│   └── repo.go                       # repo mr create / comment
└── internal/
    ├── core/config.go                # 配置加载 / 保存（~/.codearts-cli/config.json）
    ├── client/
    │   ├── signer.go                 # SDK-HMAC-SHA256 AK/SK 签名
    │   ├── client.go                 # HTTP 封装 + 三服务端点推导
    │   ├── pipeline.go               # 流水线 API
    │   ├── projectman.go             # 工作项 API
    │   └── repo.go                   # 代码托管 API
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
