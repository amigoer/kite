# Integrations

How to wire kite into the agents and runtimes you already use.

## Claude Code

The one-liner:

```bash
kite install claude
```

It writes an entry to `~/.claude.json`:

```json
{
  "mcpServers": {
    "kite": {
      "command": "/usr/local/bin/kite",
      "args": ["mcp"]
    }
  }
}
```

Restart Claude Code. Ask it: *"Create a kite room and run `echo hello` in it,
then show me the room URL."* It will call `kite_create_room`, `kite_exec`,
and report back with the URL — open it to see the live command block.

To undo: `kite uninstall claude` restores from the `.kite.bak` snapshot.

## Codex CLI

```bash
kite install codex
```

Writes to `~/.codex/config.toml`:

```toml
[mcp_servers.kite]
command = "/usr/local/bin/kite"
args = ["mcp"]
```

Behaviour mirrors the Claude install: same `.kite.bak` rollback path.

## Custom agents (HTTP)

If you write your own agent in Python, Node, Go, etc., the HTTP API is the
fastest path. Three calls cover most workflows:

```python
import requests

base = "http://127.0.0.1:8787/api/v1"

room = requests.post(f"{base}/rooms", json={"name": "py-agent"}).json()
res = requests.post(f"{base}/rooms/{room['id']}/exec", json={
    "cmd": "ls /tmp",
    "timeout_seconds": 10,
    "source": "python-agent",
}).json()
print(res["stdout"])
print("exit", res["exit_code"], "in", res["duration_ms"], "ms")
```

See [examples/python-agent/](../examples/python-agent/) for a runnable
version that also subscribes to the WebSocket stream.

## Custom agents (MCP)

If you want full MCP, point your agent's MCP client at the `kite mcp`
subprocess. The kite binary itself is the MCP server.

Generic MCP config snippet (works for most agents):

```json
{
  "command": "kite",
  "args": ["mcp", "--port", "8787"]
}
```

Tools available:

- `kite_create_room { name?, cwd? }` → `{ room_id, url }`
- `kite_exec { room_id, cmd, timeout_seconds? }` → `{ stdout, exit_code, duration_ms, truncated }`
- `kite_list_rooms {}` → `{ rooms: [...] }`
- `kite_get_room_history { room_id, limit? }` → `{ commands: [...] }`

## Operating with kite already running

If a kite daemon is already running on `:8787`, additional clients can just
talk to it; no need to launch another daemon. The CLI commands honour
`--host` and `--port`:

```bash
kite --port 8787 room list
```

## Debugging the integration

```bash
kite doctor          # checks bash, data dir, daemon, agent configs
kite serve -v        # verbose / debug logs
```
