# go-agent — a small kite-driven agent in Go

A 100-line Go program that uses kite as its shell backend to build a
project, run its tests, and report back. It demonstrates:

- using the `internal/client` SDK (mirrors what the kite CLI uses)
- orchestrating multiple commands inside one room (so cwd / env persist)
- handling exit codes and `APIError`s
- subscribing to the event stream via WebSocket for live updates

This isn't an LLM-driven agent — it's the skeleton you'd wrap one around.
The LLM decides which commands to run; this code executes them and feeds
results back.

## Run

```bash
# In one shell:
kite serve

# In another:
cd examples/go-agent
go run . /path/to/your/repo
```

You'll see something like:

```
[+] kite room: r_abc123  →  http://127.0.0.1:8787/rooms/r_abc123
[+] step 1/3: detect language
    found go.mod — Go project
[+] step 2/3: build
    ✓ go build ./... (exit 0, 412ms)
[+] step 3/3: test
    ✗ go test ./... (exit 1, 5821ms)
[+] report:
    BUILD: pass
    TEST:  FAIL (exit 1)
    See http://127.0.0.1:8787/rooms/r_abc123 for full output.
[+] room closed
```

## Files

- `main.go` — the agent loop
- `go.mod` — minimal module with one local replace
