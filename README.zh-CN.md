# kite

> 为 AI agent 而生、可被人类审计与回放的可编程 shell 会话。

[English](README.md) · **中文**

## 这是什么？

**kite** 给每个 shell 会话一个 URL，让 AI agent（Claude Code、Codex 等）
通过 HTTP / MCP API 在长生命周期的 *room* 里执行命令，与此同时人类通过
web viewer 实时审计、回放、随时接管。

kite 想解决的核心问题是：**当 agent 替你跑 shell 命令的时候，你需要随时
能回答「它到底做了什么」**——每一条命令、每一行输出、顺序、耗时、退出码。
kite 在设计上就保证了这件事：每个事件（`command.started` / `command.output`
/ `command.finished`）按 `command_id` 追加到 SQLite 日志里，任何一次跑过的
内容都可被查询、可被回放、可被审阅，永远在那里——不再丢失在滚动条里。

一个 *room* 就是 kite daemon 持有的一个常驻 `bash` 进程。它的状态（cwd、
环境变量、alias、shell 历史）跨越多次 agent 调用保持不变。同一份事件日志
同时驱动实时 viewer、回放时间轴和结构化历史 API。

## 为什么用 kite 跑 agent？

- **每一条命令都可追溯。** 每次 `exec` 产生一个稳定的 `command_id` 和三条
  追加式事件。「agent 在 14:32 干了什么？」是一条 SQL，而不是滚动条考古。
- **状态跨调用保持。** agent 可以 `cd`、`export`、source 虚拟环境、起后台
  服务——下一次调用看到的是同一个 shell 状态，不用每个 tool call 重新引导。
- **人类可以中途接管。** 任意活跃 room 都是一个真正的交互式 bash。在 web
  viewer 里点 *Take control*，或者 `kite attach <id>`，你就在 agent 正在用
  的那个 shell 里——修一下，释放，交还给 agent。
- **天然可回放。** 用时间轴拖拽 room 的任意历史，按命令文本过滤，跳到失败
  发生的那一刻。
- **MCP 原生接入。** `kite install claude`（或 `codex`）自动写入 agent 的
  MCP 配置（原配置带 `.bak` 备份）。四个工具——create room、exec、list、
  history——就能跑完整个闭环。
- **零基础设施。** SQLite、web viewer、MCP server 全部通过 `go:embed` 打包
  进单一 Go 二进制。不需要 Docker，不需要 Node.js 运行时。

## 一个典型的 agent 工作流

```text
# Agent 为一个工作单元创建 room
agent: kite_room_create(name="deploy-2026-05-17")
       → room_id = "r3kf..."

# Agent 跑命令，cwd / env 在调用之间自动保留
agent: kite_exec(r3kf, "git pull && npm ci")
       → exit_code=0, duration_ms=12830, command_id="c7a1..."

agent: kite_exec(r3kf, "npm run build")
       → exit_code=1, stderr 提示 "TS error in foo.ts:42"

# 人类用浏览器打开这个 room，看到两条命令以 block 形式排列，
# 点开失败那条看带 ANSI 颜色的完整 stderr，
# 点击 "Take control"，在活的 PTY 里直接修 bug，
# 释放控制权。Agent 重试：
agent: kite_exec(r3kf, "npm run build")
       → exit_code=0

# 之后任何人（或另一个 agent）都可以重建当时的现场：
agent: kite_history(r3kf, since=...)
       → 完整结构化记录，永久在磁盘里
```

闭环的形状很简单：**agent 驱动执行，人类审计与必要时接管，SQLite 日志是
唯一可信的来源。**

## 这不是什么

要做日常 terminal 工作的终端复用器，请用
**[Zellij](https://zellij.dev)**。kite 是互补品：它是一个面向程序化 shell
执行的 API + 事件日志，围绕 AI agent 执行场景设计。

|                   | Zellij              | kite                       |
|-------------------|---------------------|----------------------------|
| 主交互界面         | 终端                | HTTP / MCP API             |
| 数据单元           | 字节流              | 结构化命令事件             |
| 最适合的场景       | 人用终端工作        | AI agent 执行 + 审计       |
| Web viewer        | 浏览器里的终端      | 命令块 dashboard           |

## 快速开始

```bash
# 从源码构建（需要 Go 1.24+ 和 Node 20+）
git clone https://github.com/amigoer/kite.git
cd kite
make build

# 启动 daemon（前台）
./bin/kite serve

# 把 kite 接入 Claude Code 的 MCP 配置
./bin/kite install claude

# 重启 agent，让它执行：
#   「创建一个 kite room，在里面跑 `echo hello`」
# 打开 http://127.0.0.1:8787 实时观看。
```

适合脚本和一次性场景的底层 CLI 流程：

```bash
ID=$(./bin/kite room create --name demo | awk '/created/ {print $3}')
./bin/kite exec "$ID" -- echo hello
./bin/kite replay "$ID" --no-timing
./bin/kite watch "$ID"          # 浏览器里打开实时 viewer
```

## 人类自己进 room 干活

有时候你想自己驱动 room，或者从 agent 手里接管。`kite shell` 把你的终端
切到 raw 模式，键盘直接转发到 room 的 bash，体验和 `screen` 一致：

```bash
./bin/kite shell                # 创建一个 room 并直接进入
amigoer@host:~/work$ tail -f app.log
... 实时输出 ...
^C                              # 中断 tail，回到提示符（不会退出 kite）
amigoer@host:~/work$ vim README # vim / less / top 都正常
```

转义键是 `Ctrl+A`，之后按 `d` 离开（room 继续运行）、`k` 关闭 room、`?` 看
帮助、再按一次 `Ctrl+A` 则向 room 发送字面量 `Ctrl+A`。之后用
`kite attach <id>` 回到任意活跃 room，或直接 `kite attach`（不带 id）回到
最近活跃 room。room 是真正持久的 bash，cwd、env、alias 在多次 detach 间
保留。

## CLI 速查

```text
kite serve                          # 启动 daemon
kite shell                          # 创建一个 room 并进入交互式会话（类 screen）
kite attach [id]                    # 进入已有 room（不带 id 时取最近活跃的）
kite tail [id]                      # 只读跟随某个 room 的输出
kite room create [--name N]         # 创建 room
kite room list                      # 列出 room
kite room show <id>                 # 看单个 room 的详情
kite room close <id>                # 关闭 room
kite exec <id> -- <command...>      # 单次执行命令，stdout 实时输出
kite replay <id> [--speed 2.0]      # 在终端里回放事件
kite watch <id>                     # 用浏览器打开这个 room
kite web                            # 用浏览器打开 room 列表
kite install <claude|codex>         # 接入 agent 的 MCP 配置
kite uninstall <claude|codex>       # 撤销 install
kite mcp                            # 在 stdio 上启动 MCP server
kite doctor                         # 诊断当前安装
```

## 架构一段话

`kite serve` 在 `~/.kite/kite.db` 打开 SQLite，对 `~/.kite/kite.pid` 加
独占 flock，监听 `127.0.0.1:8787`。每个 room 拥有一个挂在 PTY 上的常驻
`bash --noediting --norc -i`。daemon 把命令写入这个 PTY，后面附一个 sentinel
（`__KITE_END_<exit>_<command_id>__`），这样它能在输出流里识别命令边界。
MCP server（`kite mcp`）是一个独立的 stdio 二进制，把它的工具调用——创建
room、exec、list、history——代理到 HTTP daemon。详见
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)。

## 文档

- [docs/OVERVIEW.md](docs/OVERVIEW.md)：完整的项目介绍
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)：进程模型与数据流
- [docs/PROTOCOL.md](docs/PROTOCOL.md)：HTTP / WebSocket / MCP 协议参考
- [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md)：各 agent 的接入方法

## 许可证

MIT。
