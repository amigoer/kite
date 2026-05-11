# demo — guided full feature tour

A single bash script that starts a private kite daemon and walks through
every surface kite exposes:

1. Create rooms (`build`, `test`)
2. List + show
3. Exec via HTTP API
4. Persistent shell state across execs
5. Independent rooms
6. Exit code propagation
7. Timeouts
8. Event log queries
9. Derived command summaries
10. CLI replay
11. Web viewer URLs
12. `kite doctor`
13. Cleanup

The daemon listens on a high port (default `18999`) so it won't conflict
with a daemon you already have running. Data goes to a temp dir that's
cleaned up on exit.

## Run

```bash
# From the repo root, after `make build`:
./examples/demo/full-tour.sh

# Use a different binary or port:
KITE=/usr/local/bin/kite PORT=18000 ./examples/demo/full-tour.sh
```

## Sample output

```
▸ Starting a private kite daemon on port 18999
  ✓ daemon up at http://127.0.0.1:18999 (pid 41234)

▸ 1. Creating a room via the CLI
  $ kite --port 18999 room create --name build
  room created: r_pxkt3x463mo6
  ✓ build room: r_pxkt3x463mo6
  ...

▸ 8. Querying the event log
  ✓ build room has 14 events
  first 3 event types:
    room.created
    command.started
    command.output

▸ 10. CLI replay (instant)
    $ uname -s
    Darwin
    [exit 0, 5ms]
    ...

✓ tour complete
```

## Requirements

- bash 4+
- curl
- jq
- a `kite` binary on `$PATH` or in `./bin/kite`
