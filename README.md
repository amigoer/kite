# kite

> Programmable, replayable shell sessions — built for AI agents, watchable by humans.

**English** · [中文](README.zh-CN.md)

## What is this?

**kite** gives every shell session a URL so AI agents (Claude Code, Codex,
etc.) can run commands inside long-lived *rooms* through an HTTP / MCP API —
while humans audit, replay, and intervene from a web viewer.

The problem kite solves: when an agent runs shell commands on your behalf, you
need to be able to answer "what did it actually do?" — every command, every
byte of output, the order, the duration, the exit code. kite makes that
recoverable *by design*. Every event (`command.started`, `command.output`,
`command.finished`) is appended to a SQLite log keyed by `command_id`, so any
run is queryable, replayable, and reviewable forever — not lost in scrollback.

A *room* is one long-running `bash` owned by the kite daemon. Its state — cwd,
environment, aliases, shell history — persists across agent calls. The same
event log powers the live viewer, the replay timeline, and the structured
history API.

## Why kite for AI agents?

- **Every command is traceable.** Each `exec` produces a stable `command_id`
  and three append-only events. "What did the agent do at 14:32?" is a
  single SQL query, not a scrollback dig.
- **State persists across calls.** Agents can `cd`, `export`, source venvs,
  start a background server — the next call sees that state. No fragile
  re-bootstrapping per tool call.
- **Humans can step in mid-flight.** Any active room is a real interactive
  bash. Click *Take control* in the web viewer, or run `kite attach <id>`,
  and you are inside the same shell the agent is driving — fix something,
  release, hand it back.
- **Replayable runs.** Scrub through any room's history with a timeline,
  filter by command text, jump to the exact moment a failure started.
- **MCP-native onboarding.** `kite install claude` (or `codex`) wires the
  daemon into the agent's MCP config with a backup of the original. Four
  tools — create room, exec, list, history — cover the loop.
- **Zero infrastructure.** SQLite + web viewer + MCP server are baked into
  a single Go binary via `go:embed`. No Docker, no Node.js runtime.

## A typical agent workflow

```text
# Agent creates a room for a unit of work
agent: kite_room_create(name="deploy-2026-05-17")
       → room_id = "r3kf..."

# Agent runs commands; state (cwd, env) carries between calls
agent: kite_exec(r3kf, "git pull && npm ci")
       → exit_code=0, duration_ms=12830, command_id="c7a1..."

agent: kite_exec(r3kf, "npm run build")
       → exit_code=1, stderr has "TS error in foo.ts:42"

# Human opens the room in the browser, sees both commands as blocks,
# clicks the failed one, reads the full ANSI-coloured stderr,
# clicks "Take control", types into the live PTY, fixes the bug,
# releases control. The agent retries:
agent: kite_exec(r3kf, "npm run build")
       → exit_code=0

# Later, anyone (or another agent) reconstructs the run:
agent: kite_history(r3kf, since=...)
       → full structured record on disk forever
```

The shape of the loop: **agent drives, human audits and intervenes when
needed, the SQLite log is the source of truth.**

## What this is NOT

If you want a great terminal multiplexer for daily human terminal work, use
**[Zellij](https://zellij.dev)**. kite is complementary: an API + event log
for programmatic shell sessions, shaped around AI-agent execution.

|                   | Zellij              | kite                       |
|-------------------|---------------------|----------------------------|
| Primary interface | Terminal            | HTTP / MCP API             |
| Data unit         | Byte stream         | Command events             |
| Best for          | Human terminal work | AI agent execution + audit |
| Web viewer        | Terminal in browser | Command-block dashboard    |

## Quick start

```bash
# Build from source (Go 1.24+ and Node 20+ required)
git clone https://github.com/amigoer/kite.git
cd kite
make build

# Start the daemon (foreground)
./bin/kite serve

# Wire kite into Claude Code's MCP config
./bin/kite install claude

# Restart the agent, then ask it to:
#   "Create a kite room and run `echo hello` inside it."
# Open http://127.0.0.1:8787 to watch the room in real time.
```

Lower-level CLI flows for scripts and one-shot use:

```bash
ID=$(./bin/kite room create --name demo | awk '/created/ {print $3}')
./bin/kite exec "$ID" -- echo hello
./bin/kite replay "$ID" --no-timing
./bin/kite watch "$ID"          # opens the live viewer in your browser
```

## Interactive shell (for humans)

Sometimes you want to be the one in the room — to drive it yourself, or to
take over from an agent. `kite shell` puts your terminal in raw mode talking
straight to the room's bash, like `screen`:

```bash
./bin/kite shell                # creates a room and attaches you to it
amigoer@host:~/work$ tail -f app.log
... live output ...
^C                              # interrupts tail, returns to prompt
amigoer@host:~/work$ vim README # vim, less, top all work
```

Escape is `Ctrl+A`; then `d` detaches (room keeps running), `k` closes the
room, `?` shows help, `Ctrl+A` sends a literal `Ctrl+A`. Re-enter any active
room later with `kite attach <id>`, or just `kite attach` to jump back into
the most recently active one. The room is a real persistent bash, so cwd,
env, and aliases survive between detaches.

## CLI cheat-sheet

```text
kite serve                          # start the daemon
kite shell                          # create a room and attach to it (screen-style)
kite attach [id]                    # enter an existing room (latest if id omitted)
kite tail [id]                      # read-only follow of a room's output
kite room create [--name N]         # create a room
kite room list                      # list rooms
kite room show <id>                 # one room's details
kite room close <id>                # terminate a room
kite exec <id> -- <command...>      # run one command, stream stdout
kite replay <id> [--speed 2.0]      # replay events in your terminal
kite watch <id>                     # open the room in a browser
kite web                            # open the rooms list in a browser
kite install <claude|codex>         # wire MCP config for an agent
kite uninstall <claude|codex>       # undo install
kite mcp                            # run an MCP server on stdio
kite doctor                         # diagnose your installation
```

## Architecture in one paragraph

`kite serve` opens an SQLite database under `~/.kite/kite.db`, takes an
exclusive flock on `~/.kite/kite.pid`, and listens on `127.0.0.1:8787`. Every
room owns one persistent `bash --noediting --norc -i` attached to a PTY. The
daemon writes commands to that PTY followed by a sentinel
(`__KITE_END_<exit>_<command_id>__`) so it can recognise command boundaries
in the output stream. The MCP server (`kite mcp`) is a separate stdio binary
that proxies its tool calls — create room, exec, list, history — to the
HTTP daemon. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Documentation

- [docs/OVERVIEW.md](docs/OVERVIEW.md) — the full project introduction
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — process model and data flow
- [docs/PROTOCOL.md](docs/PROTOCOL.md) — HTTP / WebSocket / MCP reference
- [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md) — agent setup recipes

## License

MIT.
