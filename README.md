# codearts-cli

华为云 **CodeArts** 命令行工具。当前覆盖三个模块共 **8 个接口**：

| 模块       | 命令                              | API                          |
| ---------- | --------------------------------- | ---------------------------- |
| 流水线     | `pipeline run`                    | RunPipeline                  |
| 流水线     | `pipeline stop`                   | StopPipelineRun              |
| 工作项管理 | `issue list`                      | ListIpdProjectIssues         |
| 工作项管理 | `issue show`                      | ShowIssueDetail              |
| 工作项管理 | `issue create`                    | CreateIpdProjectIssue        |
| 工作项管理 | `issue batch-update`              | BatchUpdateIpdIssues         |
| 代码托管   | `repo mr create`                  | CreateMergeRequest           |
| 代码托管   | `repo mr comment`                 | CreateMergeRequestDiscussion |

设计与架构参考了 `github.com/larksuite/cli`（飞书 CLI）：Cobra 命令树 +
`internal/core`（配置）+ `internal/client`（带 AK/SK 签名的 HTTP 客户端）+
`internal/output`（统一 JSON 输出）。

## 安装

推荐安装到 `~/.local/bin`（macOS/Linux 下通常已在 PATH 中，无需 sudo、无需改 shell rc）：

```bash
make install PREFIX=$HOME/.local
```

其它方式：

```bash
# 装到 /usr/local/bin（系统级，需 sudo）
sudo make install

# 只在仓库内编译，得到 ./codearts-cli
make build

# 不编译，直接跑一次
go run . --help

# 卸载
make uninstall PREFIX=$HOME/.local
```

`make install` 自身会先跑 `go build`，所以不需要先手动 `go build`。如果 `$(PREFIX)/bin` 不在
PATH 里，安装脚本会打印一行提示告诉你怎么加。

前置依赖：Go ≥ 1.23。

## 快速开始

安装完成后（假设 `codearts-cli` 已在 PATH 中）：

```bash
# 1. 配置 AK/SK（交互式，SK 不会回显）
codearts-cli config init

# 1'. 或非交互（CI 推荐 --sk-stdin，避免出现在进程列表）
echo "$HW_SK" | codearts-cli config init --ak "$HW_AK" --sk-stdin --yes

# 2. 查看配置（密钥已脱敏）
codearts-cli config show

# 3. 触发一条流水线
codearts-cli pipeline run <pipeline_id>

# 4. 停止正在运行的流水线实例
codearts-cli pipeline stop <pipeline_id> <pipeline_run_id>

# 5. 工作项
codearts-cli issue list   --issue-type US --page-no 1 --page-size 20
codearts-cli issue show   <issue_id> --issue-type US
codearts-cli issue create --title "任务标题" --description "..." \
                          --category US --assignee <user_id>
codearts-cli issue batch-update --id a,b,c --category US \
                                --attribute '{"priority":"high"}'

# 6. 创建 MR
codearts-cli repo mr create <repo_id> --title "..." --source feat/x --target main

# 7. 给 MR 发检视意见
codearts-cli repo mr comment <repo_id> <mr_iid> --body "LGTM"
```

默认租户：

| 项        | 默认值                                   |
| --------- | ---------------------------------------- |
| project_id | `cd130bd8357b4e7ab293a7979d1c8711`      |
| region    | `cn-south-1`                             |

配置文件：`~/.codearts-cli/config.json`（权限 `0600`）。端点由 region 自动推导；
特殊需要（抓包 / 私有化部署）时用环境变量覆盖，见「端点与环境变量」章节。

## 端点与环境变量

各模块的服务端点按 region 自动推导；需要时可用环境变量覆盖：

| 模块       | 默认主机                                           | 环境变量覆盖                      |
| ---------- | -------------------------------------------------- | --------------------------------- |
| 流水线     | `cloudpipeline-ext.<region>.myhuaweicloud.com`     | `CODEARTS_PIPELINE_ENDPOINT`      |
| 工作项管理 | `projectman-ext.<region>.myhuaweicloud.com`        | `CODEARTS_PROJECTMAN_ENDPOINT`    |
| 代码托管   | `codehub-ext.<region>.myhuaweicloud.com`           | `CODEARTS_REPO_ENDPOINT`          |

所有命令都支持 `--dry-run`：打印组装出的 method / path / query / body 但不发起调用，是排查签名或参数格式的首选工具。

## 命令速查

### `config`

| 子命令         | 用途                                              |
| -------------- | ------------------------------------------------- |
| `config init`  | 初始化 AK/SK、project_id、region；支持 `--sk-stdin` |
| `config show`  | 打印当前配置（AK 前四位可见，SK 全掩）            |
| `config path`  | 打印配置文件绝对路径                              |

### `pipeline run <pipeline_id>`

[`POST /v5/{project_id}/api/pipelines/{pipeline_id}/run`](https://support.huaweicloud.com/api-pipeline/RunPipeline.html)

```bash
# 使用流水线保存的默认值
codearts-cli pipeline run 7f3a...

# 指定源（分支）
codearts-cli pipeline run 7f3a... \
  --sources '[{"type":"code","params":{"build_type":"branch","target_branch":"main"}}]'

# 自定义变量
codearts-cli pipeline run 7f3a... \
  --variables '[{"name":"ENV","value":"staging"}]'

# 用文件传整个 body
codearts-cli pipeline run 7f3a... --body-file run.json
```

| Flag | 说明 |
| --- | --- |
| `--sources` / `--sources-file` | JSON 数组：按源覆盖（代码、制品…） |
| `--variables` / `--variables-file` | JSON 数组：`{name,value}` 变量 |
| `--body` / `--body-file` | 整个请求体（与上面互斥） |
| `--description` | 本次运行备注 |
| `--choose-job` / `--choose-stage` | 只跑指定 job / stage（可多次） |
| `--project-id` / `--dry-run` | 覆盖 project_id / 只预览不调用 |

### `pipeline stop <pipeline_id> <pipeline_run_id>`

[`POST /v5/{project_id}/api/pipelines/{pipeline_id}/pipeline-runs/{pipeline_run_id}/stop`](https://support.huaweicloud.com/api-pipeline/StopPipelineRun.html)

```bash
codearts-cli pipeline stop 7f3a... 816a...
```

`pipeline_run_id` 通常来自 `pipeline run` 的返回值。

### `issue list`

[`POST /v1/ipdprojectservice/projects/{project_id}/issues/query`](https://support.huaweicloud.com/api-projectman/ListIpdProjectIssues.html)

```bash
# 最简
codearts-cli issue list --issue-type US

# 分页 + 过滤
codearts-cli issue list --issue-type US,Task \
  --page-no 1 --page-size 50 \
  --filter '[{"property":"status","condition":"include","value":["new"]}]'
```

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | 单个或多个类型逗号分隔 |
| `--filter` / `--filter-file` | 过滤条件 JSON 数组 |
| `--filter-mode` | `AND_OR`（默认）/ `OR_AND` |
| `--page-no` / `--page-size` | 0 表示用 API 默认 |
| `--sort-field` / `--sort-asc` | 排序字段与方向 |

### `issue show <issue_id>`

[`GET /v1/ipdprojectservice/projects/{project_id}/issues/{issue_id}`](https://support.huaweicloud.com/api-projectman/ShowIssueDetail.html)

```bash
codearts-cli issue show 123456789 --issue-type US
```

| Flag | 说明 |
| --- | --- |
| `--issue-type`（必填） | `Epic` / `FE` / `SF` / `IR` / `RR` / `SR` / `US` / `AR` / `Bug` / `Task` |
| `--domain-id` | 可选 |

### `issue create`

[`POST /v1/ipdprojectservice/projects/{project_id}/issues`](https://support.huaweicloud.com/api-projectman/CreateIpdProjectIssue.html)

```bash
codearts-cli issue create \
  --title "接入 CodeArts CLI" \
  --description "按文档完成 AK/SK 与 RunPipeline" \
  --category US \
  --assignee <user_id_32char>
```

可选字段（`status`/`priority`）有对应 flag；需要传 `plan_iteration`、`workload_man_day`、`business_domain` 等完整字段时，用 `--body-file`。

### `issue batch-update`

[`PUT /v1/ipdprojectservice/projects/{project_id}/issues/batch`](https://support.huaweicloud.com/api-projectman/BatchUpdateIpdIssues.html)

```bash
codearts-cli issue batch-update \
  --id 111,222 --id 333 \
  --category US \
  --attribute '{"priority":"high"}'
```

`--id` 支持重复或逗号分隔。`attribute.category` 必填，通过 `--category` 或 `--attribute` 里任一方式提供。

### `repo mr create <repository_id>`

[`POST /v4/repositories/{repository_id}/merge-requests`](https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequest.html)

```bash
# 最简
codearts-cli repo mr create 12345 \
  --title "feat: 接入 codearts-cli" \
  --source feat/cli --target main

# 带评审 / 关联工作项 / squash
codearts-cli repo mr create 12345 \
  --title "feat: x" --source feat/x --target main \
  --reviewers "uid-a,uid-b" --assignees "uid-c" \
  --squash --squash-message "feat: squashed" \
  --force-remove-source \
  --work-item 1251275102548402177
```

| Flag | 说明 |
| --- | --- |
| `--title` / `--source` / `--target`（必填） | MR 标题 / 源分支 / 目标分支 |
| `--description` | MR 描述 |
| `--reviewers` / `--assignees` | 评审人 / 指派人 user_id 逗号分隔 |
| `--approval-reviewers` / `--approval-approvers` | 审批评审人 / 审批人 |
| `--work-item` | 关联工作项 ID（可重复或逗号分隔） |
| `--milestone-id` | 里程碑 ID |
| `--squash` / `--squash-message` | 合并时 squash 并指定 commit 消息 |
| `--force-remove-source` | 合并后自动删除源分支 |
| `--only-assignee-merge` | 仅允许指派人合入 |
| `--target-repo-id` | 跨仓库 MR 的目标仓库 ID |
| `--body-json` / `--body-file` | 直接传完整 JSON body（其余 flag 失效） |

`<repository_id>` 必须是整数仓库 ID。

### `repo mr comment <repository_id> <merge_request_iid>`

[`POST /v4/repositories/{repository_id}/merge-requests/{merge_request_iid}/discussions`](https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequestDiscussion.html)

```bash
# 简单评论
codearts-cli repo mr comment 12345 7 --body "请补个单测"

# 带严重级别 / 评审分类
codearts-cli repo mr comment 12345 7 \
  --body "参数未校验" --severity major --review-categories "安全"

# 行级评论（需要 position，建议用文件）
codearts-cli repo mr comment 12345 7 --body-file review.json
```

`<repository_id>` 和 `<merge_request_iid>` 都是**整数**（数字仓库 ID 与 MR iid），不是 UUID。

## 设计选择

- **不引入 huaweicloud-sdk-go-v3**：该 SDK 包含所有服务模型（几十 MB 传递依赖），
  对一个 CLI 过重。`internal/client` 直接实现 AK/SK 签名（SDK-HMAC-SHA256，算法稳定
  有公开文档），HTTP 体使用 `encoding/json`。扩展到 `projectman`、`codehub`
  等其它 CodeArts 服务时只需新增 endpoint 常量与请求方法，签名代码可复用。
- **配置落地在 `~/.codearts-cli/config.json`**：MVP 阶段直接文件存储，mode `0600`。
  飞书 CLI 把 secret 放 keychain；后续若需要可加 `internal/keychain` 层，当前先保持
  简单可移植（Linux CI 无 keychain 也能用）。
- **`--dry-run` 一等公民**：调试 AK/SK 签名、排查 pipeline 参数格式时刚需。

## 目录结构

```
codearts-cli/
├── main.go
├── Makefile                      # build / install / uninstall
├── cmd/
│   ├── root.go                   # 根命令
│   ├── config.go                 # config init / show / path
│   ├── pipeline.go               # pipeline run / stop
│   ├── issue.go                  # issue list / show / create / batch-update
│   └── repo.go                   # repo mr comment
└── internal/
    ├── core/                     # 配置加载 / 保存
    ├── client/
    │   ├── signer.go             # SDK-HMAC-SHA256 AK/SK 签名
    │   ├── client.go             # HTTP 封装 + 三服务端点推导
    │   ├── pipeline.go           # 流水线 API
    │   ├── projectman.go         # 工作项 API
    │   └── repo.go               # 代码托管 API
    └── output/                   # JSON 输出 + 成功 / 错误消息
```

## Makefile 目标

| target                                 | 用途                                             |
| -------------------------------------- | ------------------------------------------------ |
| `make build`                           | 在仓库内编译出 `./codearts-cli`                   |
| `make install`                         | 编译并安装到 `/usr/local/bin`（默认 PREFIX）      |
| `make install PREFIX=$HOME/.local`     | 安装到 `~/.local/bin`（无需 sudo）                |
| `make uninstall [PREFIX=...]`          | 从对应前缀卸载                                   |
| `make run ARGS="pipeline run <id>"`    | 编译并跑一次，`ARGS` 透传给二进制                 |
| `make vet` / `make tidy` / `make clean`| 静态检查 / 整理依赖 / 清理构建产物               |

## 扩展下一个 API

以接入 **代码托管 · 创建仓库** 为例（CodeArts Repo `CreateRepository`）：

1. 如果是新的服务域（例如制品仓 `artifact-ext`），在 `internal/client/client.go`
   加一个 `XxxEndpoint()` 方法，参考 `PipelineEndpoint` / `ProjectManEndpoint`
   / `RepoEndpoint`；接入同服务的新 API 则跳过这一步。
2. 在对应的 `internal/client/*.go` 文件里加一个方法，套路是：
   ```go
   out := map[string]interface{}{}
   if err := c.Do(ctx, "POST", c.RepoEndpoint(), path, query, body, &out); err != nil {
       return nil, err
   }
   return out, nil
   ```
3. 在 `cmd/` 下对应文件加一个 cobra 子命令，参考现有命令的 `--dry-run` 与 flag 模式。
4. 若是新模块（没有对应 `cmd/xxx.go`），建一个新文件并在 `cmd/root.go` 的
   `Execute()` 里 `root.AddCommand(newXxxCmd())`。

几个反复踩到的点（已沉淀为实现）：

- 华为 APIGW 签名要求 `CanonicalURI` **必须以 `/` 结尾**（`internal/client/signer.go:canonicalURI`）。
- 大多数 POST 即使"无参数"也必须发 `{}`，否则 `PARSE_REQUEST_DATA_EXCEPTION`
  （各 client 方法在 body=nil 时默认发 `{}`）。
- 401 `APIGW.0301` 错误会回显**服务端算出的 canonical_request**，直接和客户端版本
  diff 即可定位签名差异。
