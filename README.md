# kite

> Programmable, replayable shell sessions for AI agents and humans.

**English** · [中文](README.zh-CN.md)

## What is this?

**kite** gives every shell session a URL. AI agents (Claude Code, Codex, etc.)
execute commands inside *rooms* via an HTTP / MCP API. Humans watch in real
time through a web viewer that organizes commands into queryable, replayable
blocks.

A *room* is one long-running bash process owned by the kite daemon. Its state
— working directory, environment variables, shell history — persists across
agent calls. Every event (command started, output chunk, command finished) is
appended to an SQLite event log, which is what powers the live viewer and the
replay timeline.

## What this is NOT

If you want a great terminal multiplexer for daily terminal work, use
**[Zellij](https://zellij.dev)**. kite is complementary: an API + event log
for programmatic shell sessions, optimized for AI agent workflows.

|                   | Zellij              | kite                   |
|-------------------|---------------------|------------------------|
| Primary interface | Terminal            | HTTP / MCP API         |
| Data unit         | Byte stream         | Command events         |
| Best for          | Human terminal work | AI agent execution     |
| Web viewer        | Terminal in browser | Command-block dashboard|

## Quick start

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

## Features

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

## CLI cheat-sheet

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

## Architecture in one paragraph

`kite serve` opens an SQLite database under `~/.kite/kite.db`, takes an
exclusive flock on `~/.kite/kite.pid`, and listens on `127.0.0.1:8787`. Every
room owns one persistent `bash --noediting --norc -i` attached to a PTY. The
daemon writes commands to that PTY followed by a sentinel
(`__KITE_END_<exit>_<command_id>__`) so it can recognise command boundaries
in the output stream. The MCP server (`kite mcp`) is a separate stdio binary
that proxies all four tool calls — create room, exec, list, history — to the
HTTP daemon. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Documentation

- [docs/OVERVIEW.md](docs/OVERVIEW.md) — the full project introduction
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — process model and data flow
- [docs/PROTOCOL.md](docs/PROTOCOL.md) — HTTP / WebSocket / MCP reference
- [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md) — agent setup recipes

## License

MIT.
