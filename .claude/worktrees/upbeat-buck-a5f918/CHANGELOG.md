# Changelog

All notable changes to this project will be documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Rooms now have a `mode`: `scripted` (default — bash --norc + marker
  protocol, what kite exec / agents use) or `interactive` (the user's
  `$SHELL` launched as a normal login+interactive shell, no PS1 override,
  no startup-file skipping). `kite shell` creates interactive rooms by
  default; use `--scripted` for the old behavior. The HTTP API accepts
  `interactive: true` in the create-room body. Marker-based exec is
  rejected on interactive rooms.
- `kite shell` and `kite attach` — true screen-style interactive sessions.
  The CLI puts your terminal in raw mode and pipes bytes directly to the
  room's bash over a new `/api/v1/rooms/{id}/io` WebSocket. You get a
  native bash prompt, Ctrl+C interrupts the running command (not the
  client), Ctrl+L clears, vim/less/top work, terminal resize is forwarded
  via SIGWINCH. Escape is `Ctrl+A`; then `d` detaches, `k` closes the
  room, `?` shows help, `Ctrl+A` sends a literal `Ctrl+A`. `kite attach`
  with no id auto-resolves to the most recently active room.
- New `GET /api/v1/rooms/{id}/io` WebSocket: binary frames are raw stdin /
  stdout bytes; text frames carry JSON control messages (`{"type":"resize",
  "rows":N,"cols":N}`, `{"type":"detach"}`).
- Per-room bash mode switching: scripted (markers, echo off, no prompt) by
  default; flips to interactive (echo on, normal PS1) when any client
  attaches to `/io`, and flips back when the last detaches. Refcounted so
  multiple humans can co-attach.
- `source` field on `command.started` events now uniquely identifies which
  caller initiated a command, allowing concurrent clients to tell their own
  events apart from background activity.

### Changed

- `POST /api/v1/rooms/{id}/exec` returns HTTP 409 with code
  `interactive_attached` while any interactive client is attached to the
  room. Structured exec resumes automatically once everyone detaches.

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
