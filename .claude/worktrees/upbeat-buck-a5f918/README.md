# kite

> Programmable, replayable shell sessions for AI agents and humans.

[English](#english) · [中文](#中文)

---

## English

### What is this?

**kite** gives every shell session a URL. AI agents (Claude Code, Codex, etc.)
execute commands inside *rooms* via an HTTP / MCP API. Humans watch in real
time through a web viewer that organizes commands into queryable, replayable
blocks.

A *room* is one long-running bash process owned by the kite daemon. Its state
— working directory, environment variables, shell history — persists across
agent calls. Every event (command started, output chunk, command finished) is
appended to an SQLite event log, which is what powers the live viewer and the
replay timeline.

### What this is NOT

If you want a great terminal multiplexer for daily terminal work, use
**[Zellij](https://zellij.dev)**. kite is complementary: an API + event log
for programmatic shell sessions, optimized for AI agent workflows.

|                   | Zellij              | kite                   |
|-------------------|---------------------|------------------------|
| Primary interface | Terminal            | HTTP / MCP API         |
| Data unit         | Byte stream         | Command events         |
| Best for          | Human terminal work | AI agent execution     |
| Web viewer        | Terminal in browser | Command-block dashboard|

### Quick start

```bash
# Build from source (Go 1.24+ and Node 20+ required)
git clone https://github.com/amigoer/kite.git
cd kite
make build

# Start the daemon (foreground)
./bin/kite serve

# In another shell, wire kite into Claude Code's MCP config
./bin/kite install claude

# Restart Claude Code, then ask it to "create a kite room and run echo hello".
# Open http://127.0.0.1:8787 to watch in real time.
```

Or drop into a screen-style interactive shell — your terminal is in raw mode,
talking straight to the room's bash:

```bash
./bin/kite shell                # creates a room and attaches you to it
# inside the session, you get a normal bash prompt:
amigoer@host:~/work$ tail -f app.log
... live output ...
^C                              # interrupts tail, returns to prompt
amigoer@host:~/work$ vim README # works — vim, less, top all do
```

Escape is `Ctrl+A`; then `d` detaches (room keeps running), `k` closes the
room, `?` shows help, `Ctrl+A` sends a literal `Ctrl+A`. Re-enter any active
room later with `kite attach <id>`, or just `kite attach` to jump back into
the most recently active one. The room is a real persistent bash, so cwd,
env, and aliases survive between detaches.

Lower-level CLI flows (script-friendly, one-shot):

```bash
ID=$(./bin/kite room create --name demo | awk '/created/ {print $3}')
./bin/kite exec "$ID" -- echo hello
./bin/kite replay "$ID" --no-timing
./bin/kite watch "$ID"          # opens the live viewer in your browser
```

### Features

- **HTTP API first**: `POST /api/v1/rooms/{id}/exec` returns
  `{stdout, exit_code, duration_ms, truncated}`.
- **Commands are events**: `command.started` / `command.output` /
  `command.finished` are append-only and queryable by id.
- **Replayable**: scrub through any room's history with a timeline; filter by
  command text.
- **Zero-config agent integration**: `kite install claude` (or `codex`) writes
  the MCP server config for you, with a backup of the original.
- **Persistent shell state**: cwd, env, and aliases survive across agent calls
  inside the same room.
- **Single binary**: SQLite + web viewer + MCP server are all baked in via
  `go:embed`. No Docker, no Node.js runtime required to run kite.

### CLI cheat-sheet

```text
kite serve                          # start the daemon
kite shell                          # create a room and attach to it (screen-style)
kite attach [id]                    # enter an existing room (latest if id omitted)
kite room create [--name N]         # create a room
kite room list                      # list rooms
kite room show <id>                 # one room's details
kite room close <id>                # terminate a room
kite exec <id> -- <command...>      # run a single command, stream stdout
kite replay <id> [--speed 2.0]      # replay events in your terminal
kite watch <id>                     # open the room in a browser
kite web                            # open the rooms list in a browser
kite install <claude|codex>         # wire MCP config for an agent
kite uninstall <claude|codex>       # undo install
kite mcp                            # run an MCP server on stdio
kite doctor                         # diagnose your installation
```

Inside `kite shell` / `kite attach` it's raw bash. Escape is `Ctrl+A`,
then `d` to detach, `k` to close the room, `?` for help, `Ctrl+A` to send
a literal `Ctrl+A`.

### Architecture in one paragraph

`kite serve` opens an SQLite database under `~/.kite/kite.db`, takes an
exclusive flock on `~/.kite/kite.pid`, and listens on `127.0.0.1:8787`. Every
room owns one persistent `bash --noediting --norc -i` attached to a PTY. The
daemon writes commands to that PTY followed by a sentinel
(`__KITE_END_<exit>_<command_id>__`) so it can recognise command boundaries
in the output stream. The MCP server (`kite mcp`) is a separate stdio binary
that proxies all four tool calls — create room, exec, list, history — to the
HTTP daemon. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

### Documentation

- [docs/OVERVIEW.md](docs/OVERVIEW.md) — the full project introduction
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — process model and data flow
- [docs/PROTOCOL.md](docs/PROTOCOL.md) — HTTP / WebSocket / MCP reference
- [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md) — agent setup recipes

### License

MIT.

---

## 中文

### 这是什么？

**kite** 给每个 shell 会话一个 URL。AI agent（Claude Code、Codex 等）通过
HTTP / MCP API 在 *room* 里执行命令；人类通过 web viewer 实时观看，命令以
可查询、可回放的「命令块」呈现。

一个 *room* 就是 kite daemon 持有的一个长期运行的 bash 进程。它的状态
（cwd、环境变量、shell 历史）会跨越多次 agent 调用保持。每个事件
（command.started、command.output、command.finished）都会追加到 SQLite
事件日志，这是实时 viewer 和回放时间轴的数据基础。

### 这不是什么

要做日常 terminal 工作的终端复用器，请用
**[Zellij](https://zellij.dev)**。kite 是互补品：它是一个面向程序化 shell
执行的 API + 事件日志，主要为 AI agent 流程优化。

|                   | Zellij              | kite                   |
|-------------------|---------------------|------------------------|
| 主交互界面         | 终端                | HTTP / MCP API         |
| 数据单元           | 字节流              | 结构化命令事件          |
| 最适合的场景       | 人用终端工作        | AI agent 执行           |
| Web viewer        | 浏览器里的终端       | 命令块 dashboard        |

### 快速开始

```bash
# 从源码构建（需要 Go 1.24+ 和 Node 20+）
git clone https://github.com/amigoer/kite.git
cd kite
make build

# 启动 daemon（前台）
./bin/kite serve

# 另开一个终端，把 kite 接入 Claude Code 的 MCP 配置
./bin/kite install claude

# 重启 Claude Code，让它「创建一个 kite room 并跑 echo hello」。
# 打开 http://127.0.0.1:8787 实时观看。
```

也可以像 `screen` 一样进入一个交互式 room —— 你的终端进入 raw 模式，键盘事件
直接转发给 room 的 bash：

```bash
./bin/kite shell                # 创建一个 room 并直接进入
# 进入后是正常的 bash 提示符：
amigoer@host:~/work$ tail -f app.log
... 实时输出 ...
^C                              # 中断 tail，回到提示符（不会退出 kite）
amigoer@host:~/work$ vim README # 可以用，vim/less/top 都正常
```

转义键是 `Ctrl+A`，之后按 `d` 离开（room 继续运行）、`k` 关闭 room、`?` 查看
帮助、再按一次 `Ctrl+A` 则向 room 发送字面量 `Ctrl+A`。之后用
`kite attach <id>` 回到任意活跃 room，或直接 `kite attach`（不带 id）回到最近
活跃 room。room 是一个真正持久的 bash，cwd、env、alias 在多次 detach 之间保留。

底层 CLI 流程（适合脚本、单条命令场景）：

```bash
ID=$(./bin/kite room create --name demo | awk '/created/ {print $3}')
./bin/kite exec "$ID" -- echo hello
./bin/kite replay "$ID" --no-timing
./bin/kite watch "$ID"          # 浏览器里打开实时 viewer
```

### 核心特性

- **API 是第一公民**：`POST /api/v1/rooms/{id}/exec` 返回
  `{stdout, exit_code, duration_ms, truncated}`。
- **命令是事件**：`command.started` / `command.output` /
  `command.finished` 追加写入，可按 id 查询。
- **天然可回放**：用时间轴拖拽 room 的任意历史；可按命令文本过滤。
- **agent 零配置接入**：`kite install claude`（或 `codex`）自动写入 MCP
  配置，原配置带 `.bak` 备份。
- **持久 shell 状态**：cwd、env、alias 在同一个 room 内跨越多次 agent 调用
  保持。
- **单二进制**：SQLite、web viewer、MCP server 全部通过 `go:embed` 打包，
  不需要 Docker，不需要 Node.js 运行时。

### CLI 速查

```text
kite serve                          # 启动 daemon
kite shell                          # 创建一个 room 并进入交互式会话（类 screen）
kite attach [id]                    # 进入已有 room（不带 id 时取最近活跃的）
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

`kite shell` / `kite attach` 内部是原生 bash。转义键 `Ctrl+A`，之后 `d` 离开、
`k` 关闭 room、`?` 帮助，再按 `Ctrl+A` 发送字面量 `Ctrl+A`。

### 架构一段话

`kite serve` 在 `~/.kite/kite.db` 打开 SQLite，对 `~/.kite/kite.pid` 加
独占 flock，监听 `127.0.0.1:8787`。每个 room 拥有一个挂在 PTY 上的常驻
`bash --noediting --norc -i`。daemon 把命令写入这个 PTY，后面附一个 sentinel
（`__KITE_END_<exit>_<command_id>__`），这样它能在输出流里识别命令边界。
MCP server（`kite mcp`）是一个独立的 stdio 二进制，把四个工具调用——创建
room、exec、list、history——代理到 HTTP daemon。详见
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)。

### 文档

- [docs/OVERVIEW.md](docs/OVERVIEW.md)：完整的项目介绍
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)：进程模型与数据流
- [docs/PROTOCOL.md](docs/PROTOCOL.md)：HTTP / WebSocket / MCP 协议参考
- [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md)：各 agent 的接入方法

### 许可证

MIT。
