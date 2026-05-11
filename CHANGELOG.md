# Changelog

All notable changes to this project will be documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `kite shell` and `kite attach` — screen-style interactive room sessions.
  Enter a room once, type commands at a `kite (room_id)>` prompt, and detach
  with `Ctrl+D` or close with `:close`. Meta commands (`:help`, `:detach`,
  `:close`, `:status`, `:url`, `:history`, `:clear`) start with `:`. `kite
  attach` with no id auto-resolves to the most recently active room.
- `source` field on `command.started` events now uniquely identifies which
  caller initiated a command, allowing concurrent clients to tell their own
  events apart from background activity.

### Fixed

- Extra trailing blank line after each command's output in the PTY stream. The
  marker protocol's leading `\n` is now folded into the marker boundary
  instead of being emitted as user-visible output.

## [0.1.0] — 2026-05-12

### Added

- HTTP API at `127.0.0.1:8787` for room CRUD, command execution, event
  queries, and command-history summaries.
- WebSocket stream (`/api/v1/rooms/{id}/stream`) that pushes `init` and
  per-event messages with heartbeats.
- Persistent bash sessions per room, with the
  `__KITE_END_<exit>_<command_id>__` marker protocol for command boundaries.
- SQLite-backed event log (`~/.kite/kite.db`) in WAL mode.
- CLI: `serve`, `room create|list|show|close`, `exec`, `watch`, `web`,
  `replay`, `install`, `uninstall`, `mcp`, `doctor`, `version`.
- MCP server on stdio with four tools: `kite_create_room`, `kite_exec`,
  `kite_list_rooms`, `kite_get_room_history`.
- One-click installers for Claude Code (`~/.claude.json`) and Codex
  (`~/.codex/config.toml`), with `.kite.bak` rollback.
- Embedded web viewer (Vite + vanilla TypeScript) with command-block
  dashboard, live mode, and replay mode (timeline scrub, speed, search).
- Architecture, protocol, and integrations docs; examples for curl, Python,
  and Claude Code.
- Release pipeline via goreleaser (linux/darwin × amd64/arm64), Homebrew
  tap, and `install.sh` script.

### Known limitations

- stdout and stderr are merged into one `stream: "stdout"`. Full separation
  is planned for 0.2.
- Marker collision with user-typed text containing the exact pattern is
  unlikely but possible; 0.2 will use a per-room random marker.
- Windows is not supported (Linux + macOS only).

[0.1.0]: https://github.com/amigoer/kite/releases/tag/v0.1.0
