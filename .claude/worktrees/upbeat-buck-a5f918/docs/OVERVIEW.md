# kite — Project Overview / 项目概览

[English](#english) · [中文](#中文)

---

## English

### The problem in one sentence

AI agents run shell commands constantly, but the moment a command finishes
its output disappears into a chat transcript — non-queryable, non-replayable,
non-observable. kite gives every command a permanent home: a URL, a row in a
database, an event in a stream.

### Why now

Two trends collided:

1. **AI agents are doing real shell work.** Claude Code, Codex, Aider, Cursor
   Agent — they all execute shell commands as part of their reasoning. A
   medium-complexity coding task can run 50–100 commands.
2. **Humans need to supervise them.** Reviewing what an agent did means
   re-reading a noisy chat log. There's no scrub bar. There's no per-command
   exit code. There's no way to share "look at command 47" with a teammate.

Existing tools don't fit:

- **tmux / Zellij** record bytes, not commands. You can't say "show me the
  command that exited 1 and lasted 2.3 seconds."
- **asciinema** records the screen, not the structure. It's a player, not
  an API.
- **Plain log files** lose timing, ordering between concurrent sessions,
  exit codes, command boundaries.

kite picks a different primitive: **commands are events**, and rooms are
event streams.

### The mental model

Three concepts, that's it.

#### Room

A room is one persistent bash process owned by the kite daemon. You create
one, point an agent (or a human, or both) at it, and it lives until you
close it. Its shell state — `cd`, `export`, aliases, even an open
`virtualenv` — persists across every command anyone runs.

```
Room r_abc123
├─ bash (PID 4242, cwd: /repo, env: …)
├─ command c_xyz       echo hello              → exit 0, 12ms
├─ command c_xyz2      git status              → exit 0, 89ms
└─ command c_xyz3      go build ./...          → exit 2, 5421ms
```

A room can have multiple participants (agents, humans, CLI users). They
all share the same shell state — kite serialises their `exec` calls under
the hood so the marker protocol stays unambiguous.

#### Event

Every state change writes one row to SQLite. Seven event types cover the
whole lifecycle:

| Type                  | Emitted when                                          |
|-----------------------|-------------------------------------------------------|
| `room.created`        | A new room comes up                                   |
| `room.closed`         | The bash dies or someone closes the room              |
| `command.started`     | An `exec` call begins                                 |
| `command.output`      | A chunk of stdout arrives from the PTY                |
| `command.finished`    | The boundary marker is seen, with exit code + duration|
| `participant.joined`  | (reserved for 0.2)                                    |
| `participant.left`    | (reserved for 0.2)                                    |

Everything kite shows you — the live viewer, the replay timeline, the
`commands` summary — is derived from this stream. Add a new view? Just
write another reducer over the event log.

#### Marker protocol

This is the only non-obvious bit. bash doesn't tell you "command 5 just
finished." kite makes it tell you, by appending a sentinel after every
command:

```sh
ls -la
printf '\n__KITE_END_%d_c_xyz__\n' $?
```

The PTY reader looks for `__KITE_END_<exit>_<command_id>__`. Output before
the marker belongs to that command; the marker line is filtered out before
it reaches viewers. Trivial idea, surprising amount of plumbing — see
[ARCHITECTURE.md](ARCHITECTURE.md) for the corner cases.

### A tour through the surfaces

kite gives you five ways to do the same thing. Pick what fits the situation.

#### Interactive shell (the human-facing default)

```bash
kite shell                 # creates a room and drops you into it
# inside:
kite (r_abc123)> go test ./...
kite (r_abc123)> :history
kite (r_abc123)> :detach   # leave it running; come back later with `kite attach`
```

The same room can be re-entered any time with `kite attach r_abc123`, or with
just `kite attach` to jump back into the most recently active one. Output
streams over the room's WebSocket, so anything an agent runs in the same
room shows up live in your terminal too.

#### CLI (script-friendly, one-shot)

```bash
kite serve                           # start the daemon
kite room create --name build        # r_abc123
kite exec r_abc123 -- go test ./...  # stream output, return exit code
kite replay r_abc123                 # play back the whole session
```

#### HTTP API

```bash
curl -s -X POST http://127.0.0.1:8787/api/v1/rooms/r_abc123/exec \
  -H 'Content-Type: application/json' \
  -d '{"cmd":"git status","timeout_seconds":5}' | jq
# {
#   "command_id": "c_xyz",
#   "stdout": "On branch main\nnothing to commit\n",
#   "exit_code": 0,
#   "duration_ms": 89,
#   "truncated": false
# }
```

Full reference: [PROTOCOL.md](PROTOCOL.md).

#### MCP (for AI agents)

```bash
kite install claude    # writes ~/.claude.json
# restart Claude Code, then ask it:
#   "Create a kite room, build the project, and tell me if tests pass."
```

The agent calls `kite_create_room` + `kite_exec` internally; the human
watches in the browser.

#### Web viewer

`http://127.0.0.1:8787/rooms/r_abc123` opens a dark, monospace dashboard
where every command is a collapsible block. Live mode follows the WebSocket
stream; replay mode lets you scrub the timeline and filter by command
text.

### Compared to…

| Tool            | Records bytes? | Records commands? | API? | Replay? | Agent-friendly? |
|-----------------|----------------|-------------------|------|---------|-----------------|
| tmux / Zellij   | ✓              | ✗                 | ✗    | ✗       | ✗               |
| script(1)       | ✓              | ✗                 | ✗    | partial | ✗               |
| asciinema       | ✓              | ✗                 | ✗    | ✓       | ✗               |
| shell history   | ✗              | partial           | ✗    | ✗       | ✗               |
| **kite**        | ✓              | **✓**             | **HTTP+MCP** | **✓** | **✓**     |

This isn't a takedown; tmux and asciinema are great at what they do. kite
just makes a different trade: structured commands over byte fidelity, API
access over interactive UX.

### When to use kite

- You're building or running AI agents that touch the shell.
- You want to see what an agent did *as it does it*, without piping every
  command into your chat client.
- You want a record of what ran, with exit codes, durations, and ordering.
- You want to replay or share an agent's session.
- You want a thin programmatic shell layer in your own tool.

### When NOT to use kite

- You want a daily-driver terminal multiplexer → use Zellij or tmux.
- You want a terminal emulator → kite isn't one and never will be.
- You need byte-perfect ANSI capture for screen recordings → use asciinema.
- You need to run untrusted code in a sandbox → kite is bash, not a
  container. Wrap it in your own isolation.

### What's in v0.1 (and what isn't)

In scope:

- HTTP + WebSocket API on 127.0.0.1
- Persistent bash session per room, with the marker protocol
- SQLite event log with WAL
- CLI: serve / shell / attach / room / exec / watch / replay / install / doctor / mcp
- MCP stdio server with 4 tools
- One-click MCP installers for Claude Code + Codex
- Web viewer (Live + Replay)
- Linux + macOS, single binary

Out of scope (explicit non-goals for v0.1):

- Multi-pane / window / layout
- Terminal emulator UI in the browser
- Plugins
- Policy engine / command approval
- Multi-writer conflict UI (we serialize via mutex)
- Federation / remote rooms
- Auth / tokens (loopback only)
- Windows
- TUI viewer

### Where to go next

| If you want to…                       | Read                                  |
|---------------------------------------|---------------------------------------|
| Build kite from source                | [../README.md](../README.md)          |
| Understand the runtime                | [ARCHITECTURE.md](ARCHITECTURE.md)    |
| Call the HTTP / WS / MCP API          | [PROTOCOL.md](PROTOCOL.md)            |
| Wire kite into a specific agent       | [INTEGRATIONS.md](INTEGRATIONS.md)    |
| See a runnable script                 | [../examples/demo/full-tour.sh](../examples/demo/full-tour.sh) |
| See a Go-based agent skeleton         | [../examples/go-agent/](../examples/go-agent/) |

---

## 中文

### 一句话讲清楚问题

AI agent 不停地在跑 shell 命令，但命令一执行完，输出就消失进聊天记录里——
不能查询、不能回放、不能观察。kite 给每条命令一个永久的家：一个 URL、
数据库里的一行、事件流里的一个事件。

### 为什么是现在

两个趋势撞在一起：

1. **AI agent 在做真实的 shell 工作**。Claude Code、Codex、Aider、Cursor
   Agent——它们都把执行 shell 命令作为推理的一部分。一个中等复杂度的
   编码任务可能跑 50–100 条命令。
2. **人类需要监督它们**。复盘一个 agent 做了什么，意味着回头读一长串
   嘈杂的聊天记录。没有进度条。没有按命令的退出码。没有办法跟同事说
   「看第 47 条命令」。

现有工具都不合适：

- **tmux / Zellij** 录的是字节，不是命令。你没法说「给我看那条退出码是 1、
  跑了 2.3 秒的命令」。
- **asciinema** 录的是屏幕，不是结构。它是个播放器，不是个 API。
- **裸 log 文件** 丢失时序、并发会话之间的顺序、退出码、命令边界。

kite 选了一个不同的原语：**命令就是事件**，room 就是事件流。

### 心智模型

三个概念，仅此而已。

#### Room（房间）

一个 room 就是 kite daemon 拥有的一个持久化 bash 进程。你创建一个，
让 agent（或人、或两者）连上去，它会一直活到你显式关闭它。它的 shell
状态——`cd`、`export`、alias，甚至打开的 `virtualenv`——跨越每个人跑的
每条命令保持。

```
Room r_abc123
├─ bash (PID 4242, cwd: /repo, env: …)
├─ command c_xyz       echo hello              → exit 0, 12ms
├─ command c_xyz2      git status              → exit 0, 89ms
└─ command c_xyz3      go build ./...          → exit 2, 5421ms
```

一个 room 可以有多个参与者（agent、人、CLI 用户），他们共享同一份 shell
状态——kite 在底层串行化他们的 `exec` 调用，让 marker 协议保持无歧义。

#### Event（事件）

每个状态变化都往 SQLite 写一行。七种事件类型涵盖整个生命周期：

| 类型                   | 触发时机                                              |
|-----------------------|-------------------------------------------------------|
| `room.created`        | 新 room 创建                                          |
| `room.closed`         | bash 退出或 room 被关闭                                |
| `command.started`     | 一次 `exec` 开始                                       |
| `command.output`      | PTY 输出一个数据块                                     |
| `command.finished`    | 看到边界 marker，附带退出码 + 时长                       |
| `participant.joined`  | （0.2 预留）                                          |
| `participant.left`    | （0.2 预留）                                          |

kite 给你看的所有东西——实时 viewer、回放时间轴、`commands` 摘要——都
是从这个事件流推导出来的。要加个新视图？写一个事件 reducer 就行了。

#### Marker 协议

这是唯一不太显然的点。bash 不会主动告诉你「第 5 条命令刚跑完」。kite 让
它告诉你，方法是每条命令后面接一个 sentinel：

```sh
ls -la
printf '\n__KITE_END_%d_c_xyz__\n' $?
```

PTY reader 扫描 `__KITE_END_<exit>_<command_id>__`。marker 之前的输出归
属那条命令；marker 本身在到达 viewer 之前被过滤掉。想法很简单，工程量
意外地大——细节见 [ARCHITECTURE.md](ARCHITECTURE.md)。

### 五种使用方式

kite 提供五种途径做同一件事。按场景选。

#### 交互式 shell（推荐给人用）

```bash
kite shell                 # 创建一个 room 并直接进入
# 在 room 里面：
kite (r_abc123)> go test ./...
kite (r_abc123)> :history
kite (r_abc123)> :detach   # 离开但保持运行；之后用 `kite attach` 回来
```

之后可以用 `kite attach r_abc123` 随时回到同一个 room，或者直接
`kite attach`（不带 id）回到最近活跃的那个。输出走的是 room 的
WebSocket，所以 agent 在同一个 room 里跑的东西也会在你的终端里实时显示。

#### CLI（适合脚本、单次命令）

```bash
kite serve                           # 启动 daemon
kite room create --name build        # r_abc123
kite exec r_abc123 -- go test ./...  # 流式输出，返回 exit code
kite replay r_abc123                 # 回放整个会话
```

#### HTTP API

```bash
curl -s -X POST http://127.0.0.1:8787/api/v1/rooms/r_abc123/exec \
  -H 'Content-Type: application/json' \
  -d '{"cmd":"git status","timeout_seconds":5}' | jq
```

完整参考：[PROTOCOL.md](PROTOCOL.md)。

#### MCP（给 AI agent 用）

```bash
kite install claude    # 写入 ~/.claude.json
# 重启 Claude Code，跟它说：
#   「创建一个 kite room，build 项目，告诉我测试是否通过。」
```

agent 内部调 `kite_create_room` + `kite_exec`；人在浏览器里看。

#### Web viewer

`http://127.0.0.1:8787/rooms/r_abc123` 打开一个深色等宽 dashboard，每
条命令是一个可折叠的块。Live 模式跟随 WebSocket；Replay 模式让你拖
时间轴、按命令文本过滤。

### 与其他工具的对比

| 工具            | 录字节？  | 录命令？  | API？ | 回放？  | agent 友好？    |
|-----------------|-----------|-----------|-------|---------|-----------------|
| tmux / Zellij   | ✓         | ✗         | ✗     | ✗       | ✗               |
| script(1)       | ✓         | ✗         | ✗     | 部分    | ✗               |
| asciinema       | ✓         | ✗         | ✗     | ✓       | ✗               |
| shell history   | ✗         | 部分      | ✗     | ✗       | ✗               |
| **kite**        | ✓         | **✓**     | **HTTP+MCP** | **✓** | **✓**       |

不是说前者不好；tmux 和 asciinema 各自的场景都做得很好。kite 只是选了
不同的 trade-off：结构化命令优于字节级保真，API 访问优于交互式 UX。

### 什么时候用 kite

- 你在做或在跑会接触 shell 的 AI agent。
- 你想在 agent 做事时实时看它做什么，而不是把每条命令塞进 chat。
- 你想要一份 ran-what 记录，附带退出码、时长、顺序。
- 你想回放或分享一次 agent 会话。
- 你想在自己的工具里加一层薄薄的程序化 shell。

### 什么时候不用 kite

- 你想要日常用的终端复用器 → 用 Zellij 或 tmux。
- 你想要一个终端模拟器 → kite 不是、也永远不会是。
- 你想要字节级 ANSI 捕获做屏幕录像 → 用 asciinema。
- 你需要在沙箱里跑不可信代码 → kite 就是 bash，不是容器。请自己加隔离。

### v0.1 范围

包含：

- 监听在 127.0.0.1 的 HTTP + WebSocket API
- 每个 room 一个持久 bash 进程 + marker 协议
- WAL 模式的 SQLite 事件日志
- CLI：serve / shell / attach / room / exec / watch / replay / install / doctor / mcp
- 提供 4 个工具的 MCP stdio server
- Claude Code + Codex 的一键 MCP 安装器
- Web viewer（Live + Replay）
- Linux + macOS，单二进制

明确不做：

- 多 pane / window / layout
- 浏览器里的终端模拟器 UI
- 插件系统
- 策略引擎 / 命令审批
- 多写者冲突 UI（我们用 mutex 串行化）
- Federation / 跨机器 room
- Auth / token（仅 loopback）
- Windows
- TUI viewer

### 接下来该看哪里

| 如果你想……                            | 看                                      |
|---------------------------------------|---------------------------------------|
| 从源码构建 kite                        | [../README.md](../README.md)          |
| 理解运行时                             | [ARCHITECTURE.md](ARCHITECTURE.md)    |
| 调用 HTTP / WS / MCP API              | [PROTOCOL.md](PROTOCOL.md)            |
| 接入某个具体 agent                     | [INTEGRATIONS.md](INTEGRATIONS.md)    |
| 跑一个完整脚本                         | [../examples/demo/full-tour.sh](../examples/demo/full-tour.sh) |
| 看 Go 写的 agent 骨架                  | [../examples/go-agent/](../examples/go-agent/) |
