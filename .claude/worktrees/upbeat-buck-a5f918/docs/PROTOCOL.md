# Protocol Reference

All endpoints are prefixed `/api/v1`. The daemon binds to `127.0.0.1:8787`.

## Error envelope

Every error response uses the same shape:

```json
{
  "error": {
    "code": "room_not_found",
    "message": "Room r_xxx does not exist"
  }
}
```

Status codes:

| Code | Meaning                                          |
|------|--------------------------------------------------|
| 400  | Malformed request body                           |
| 404  | Room not found                                   |
| 408  | Command exceeded `timeout_seconds`               |
| 409  | Operation invalid for the room's current status (e.g. closed) |
| 500  | Internal error (check daemon logs)               |

## Room endpoints

### `POST /api/v1/rooms`

Create a new room. All body fields are optional.

```json
{ "name": "release-tests", "cwd": "/tmp", "shell": "/bin/bash" }
```

Response (200):

```json
{
  "id": "r_abc123xyz789",
  "name": "release-tests",
  "created_at": "2026-05-12T02:07:42Z",
  "status": "active",
  "cwd": "/tmp",
  "shell": "/bin/bash",
  "url": "/rooms/r_abc123xyz789",
  "command_count": 0
}
```

### `GET /api/v1/rooms?status=active&limit=50`

```json
{ "rooms": [ /* room objects */ ] }
```

### `GET /api/v1/rooms/{id}`

Same shape as create response. Returns 404 if missing.

### `DELETE /api/v1/rooms/{id}`

Terminates the bash process and appends a `room.closed` event.

```json
{ "status": "closed" }
```

## Command execution

### `POST /api/v1/rooms/{id}/exec`

```json
{ "cmd": "ls -la", "timeout_seconds": 60, "source": "api" }
```

`source` is recorded into `command.started.payload.source` for traceability.
`timeout_seconds` triggers SIGINT (Ctrl+C) via the PTY when exceeded.

Response (200):

```json
{
  "command_id": "c_xyz",
  "stdout": "...",
  "exit_code": 0,
  "duration_ms": 234,
  "truncated": false
}
```

`truncated` becomes true when the output exceeded the daemon's per-command
cap (8 MiB by default). The full bytes are still in the event log, just not
in the response.

## Event log

### `GET /api/v1/rooms/{id}/events?after_id=100&limit=200&type=command.started`

```json
{ "events": [ ... ], "next_after_id": 300 }
```

`next_after_id` lets you long-poll: pass it back as `after_id` to fetch
anything newer.

Event shape:

```json
{
  "id": 7,
  "room_id": "r_abc",
  "timestamp": "2026-05-12T02:07:42Z",
  "type": "command.finished",
  "payload": { "command_id": "c_xyz", "exit_code": 0, "duration_ms": 12 }
}
```

### `GET /api/v1/rooms/{id}/commands`

Derived view: one entry per command, with output size and exit code already
folded in.

```json
{
  "commands": [
    {
      "command_id": "c_xyz",
      "cmd": "ls",
      "source": "api",
      "started_at": "2026-05-12T02:07:42Z",
      "finished_at": "2026-05-12T02:07:42.234Z",
      "exit_code": 0,
      "duration_ms": 234,
      "output_size": 1234
    }
  ]
}
```

## WebSocket stream

`ws://127.0.0.1:8787/api/v1/rooms/{id}/stream`

On connect the server sends one `init` message:

```json
{ "type": "init", "room": { ... }, "recent_events": [ /* last 100 */ ] }
```

After that, every appended event for this room is pushed as:

```json
{ "type": "event", "event": { /* same shape as the REST endpoint */ } }
```

The server pings every 25 seconds. Clients should reconnect on close (the
reference web viewer uses a 1.5s backoff).

## Health

`GET /healthz` → `{ "status": "ok", "version": "0.1.0" }`

## MCP tools

The MCP server (`kite mcp`, served over stdio) exposes four tools. Each
forwards to the HTTP API of a running daemon:

| Tool                       | Maps to                                   |
|----------------------------|-------------------------------------------|
| `kite_create_room`         | `POST /api/v1/rooms`                      |
| `kite_exec`                | `POST /api/v1/rooms/{id}/exec`            |
| `kite_list_rooms`          | `GET  /api/v1/rooms`                      |
| `kite_get_room_history`    | `GET  /api/v1/rooms/{id}/commands`        |

All tool results are JSON-encoded text content.

## Marker format

bash boundary marker (informational; emitted internally, never visible to
clients):

```
__KITE_END_<exit_code>_<command_id>__
```

Where `command_id` matches the regex `c_[a-z2-7]{12}` (a base32 lowercase
suffix).
