# Architecture

```
                ┌────────────────────────────────────────────────────┐
                │                    kite daemon                     │
                │                                                    │
   HTTP / WS    │  ┌──────────┐    ┌──────────┐    ┌──────────────┐  │
   ───────────► │  │  server  ├───►│   room   ├───►│ pty.Session  │──┼──► bash
   127.0.0.1   │  │ (net/http│    │ Manager  │    │ (creack/pty) │  │
                │  │   1.22+) │    │          │    │  + marker    │  │
                │  └────┬─────┘    └────┬─────┘    └──────┬───────┘  │
                │       │               │                 │          │
                │       │          ┌────▼─────┐           │          │
                │       └─────────►│  store   │◄──────────┘          │
                │                  │ (SQLite, │                      │
                │                  │   WAL)   │                      │
                │                  └────┬─────┘                      │
                │                       │ pubsub                     │
                │                  ┌────▼─────┐                      │
                │                  │ WS /stream│                     │
                │                  └──────────┘                      │
                └────────────────────────────────────────────────────┘

  Agents:                                       Humans:
  ─────────────────                              ────────
  kite mcp (stdio)  ◄── HTTP ───── kite serve  ◄── browser
                                                  / web/dist
```

## Processes and lifetimes

| Process            | Lifetime                                |
|--------------------|-----------------------------------------|
| `kite serve`       | Long-running daemon, one per host       |
| `kite mcp`         | Spawned by agent; one per agent process |
| bash inside a room | Lives until the room is closed          |
| Client (CLI / agent / browser) | Comes and goes; rooms persist independently |

`exec.CommandContext` is intentionally NOT used for the bash process. The
session lifetime is controlled by `pty.Session.Close()` only — an HTTP
request finishing must not kill bash.

## The marker protocol

bash never tells you "command N just finished" out of the box. We synthesise
that signal by appending a sentinel after every command we write:

```sh
ls -la
printf '\n__KITE_END_%d_c_abc123xyz789__\n' $?
```

The reader scans bytes for `__KITE_END_<exit>_<command_id>__`. Everything
before that marker belongs to that command's stdout; the marker itself is
filtered out before being broadcast to viewers.

To keep output clean we:

- launch bash with `--noediting --norc -i`,
- pass `PS1=`, `PS2=`, `PROMPT_COMMAND=`, `TERM=dumb` in env,
- run `stty -echo -onlcr` as the first synthesised command (the bootstrap
  exec; output is discarded).

Known limitations:

- A command whose stdout literally contains `__KITE_END_<digits>_c_…__`
  would confuse the parser. v0.2 will use a per-room random marker.
- stdout and stderr are merged into one stream (`stream: "stdout"`). Full
  stderr separation requires a second named pipe; deferred to v0.2.

## The event log

Every state change is an event appended to the `events` table. The room's
"current state" is derivable from its event stream, which is what makes
replay a free feature rather than a separate code path.

Event types:

- `room.created` / `room.closed`
- `command.started` / `command.output` / `command.finished`
- `participant.joined` / `participant.left` (reserved; not emitted in v0.1)

Each output chunk becomes one `command.output` event with the raw bytes
inside `payload.data` (base64-encoded over JSON). The viewer's command-block
renders these in order; the replay timeline lets you scrub to any cutoff.

## Concurrency model

- One process-wide lock: the daemon's PID flock at `~/.kite/kite.pid`.
- One mutex per room (`roomSession.execMu`) serialises Exec calls so the
  marker protocol stays unambiguous. Multiple callers queue — they don't
  get rejected.
- The `store.bus` broadcasts every appended event to all subscribers
  (typically WebSocket viewers). Slow subscribers drop events; they can
  recover via `GET /events?after_id=…`.

## Data layout

```
~/.kite/
├── kite.db            ─ SQLite file (WAL mode)
├── kite.db-wal        ─ write-ahead log
├── kite.db-shm        ─ shared memory
└── kite.pid           ─ flock'd PID file
```

Override the location with `$KITE_HOME` (preferred) or `$XDG_DATA_HOME/kite`.
