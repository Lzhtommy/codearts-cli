---
name: codearts-shared
version: 0.1.1
description: "CodeArts CLI 共享基础：AK/SK 配置初始化、凭证管理、通用标志（--dry-run）、端点解析、错误处理。首次使用 codearts-cli 或遇到认证/配置问题时触发。"
metadata:
  category: "devops"
  requires:
    bins: ["codearts-cli"]
  cliHelp: "codearts-cli --help"
---

# codearts-shared (v1)

华为云 CodeArts CLI 的共享基础知识。**所有其它 codearts-* skill 都依赖本 skill**——在使用 pipeline / issue / repo 命令前，必须先完成配置。

## 安装

```bash
# npm 全局安装（自动下载预编译二进制）
npm install -g @autelrobotics/codearts-cli

# 或从源码
git clone https://github.com/Lzhtommy/codearts-cli.git
cd codearts-cli && make install PREFIX=$HOME/.local
```

## 配置初始化

```bash
# 交互式（推荐，SK 不会回显）
codearts-cli config init
```

交互式会依次询问 5 项：

| # | 字段       | 必填 | 说明                                              |
| - | ---------- | ---- | ------------------------------------------------- |
| 1 | AK         | 是   | IAM Access Key ID                                 |
| 2 | SK         | 是   | IAM Secret Access Key（输入时不回显）             |
| 3 | Project ID | 默认 | CodeArts 项目 UUID（工作项接口直接使用此值，流水线/repo 可显式覆盖） |
| 4 | Gateway    | 默认 | CodeArts 网关 URL，默认 `http://10.250.63.100:8099` |
| 5 | User ID    | 可选 | IAM user_id（32 位 UUID），issue create 默认 assignee |

### 非交互 / CI 模式

```bash
# SK 通过 stdin 传入（推荐：不暴露在进程列表）
echo "$HW_SK" | codearts-cli config init \
    --ak "$HW_AK" --sk-stdin \
    --project-id <uuid> --user-id <uuid> --yes

# 修改单个字段（不用走完整 init）
codearts-cli config set userId <uuid>
codearts-cli config set gateway http://<host>:<port>
```

### 配置文件

- 路径：`~/.codearts-cli/config.json`（权限 `0600`）
- 查看：`codearts-cli config show`（AK 脱敏、SK 全掩）
- 定位：`codearts-cli config path`

## 通用标志

| 标志           | 说明                                          |
| -------------- | --------------------------------------------- |
| `--project-id` | 流水线 / `repo list` 命令必填；工作项命令**不支持**，统一用 config 的 `projectId` |
| `--dry-run`    | 所有命令都支持：打印 method/path/body 但不发请求 |

## 端点解析

所有服务共用 `config.json` 的 `gateway` 字段，默认 `http://10.250.63.100:8099`。不再按 region 推导，不再支持 `CODEARTS_*_ENDPOINT` 环境变量。切换网关：

```bash
codearts-cli config set gateway http://<host>:<port>
```

## 错误处理

### 401 APIGW.0301 签名失败

服务端会在错误体中回显它自己算出的 `canonical_request`，直接与客户端的对比即可定位差异。常见原因：
- AK/SK 错误 → `codearts-cli config show` 检查
- 网关 URL 错误 → `codearts-cli config set gateway <correct-url>`

### 400 PARSE_REQUEST_DATA_EXCEPTION

POST 请求没有 body。codearts-cli 已自动发 `{}`，通常不会遇到。

### PM.02177003 非目标项目成员

`--assignee` 里的 user_id 不是该 project 的成员。注意区分：
- `tenant_id / domain_id`（租户 ID）
- `user_id`（用户 ID）

两者都是 32 位 hex 但含义不同。正确的 user_id 获取路径：华为云控制台 → 我的凭证 → API 凭证 → IAM 用户 ID。

## 安全规则

- **不要** 在消息、日志、issue 描述中泄露 AK/SK
- `config show` 已自动脱敏
- CI 场景用 `--sk-stdin` 避免 SK 出现在进程列表
