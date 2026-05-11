# claude-code example

Walkthrough for wiring kite into Claude Code (Anthropic's CLI).

## One-time setup

```bash
kite install claude
```

This writes an entry to `~/.claude.json`:

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

The original file is backed up to `~/.claude.json.kite.bak`. Restart Claude
Code so it re-reads the config.

## Try it

Inside Claude Code, type:

> Create a kite room, run `uname -a` and `df -h` in it, then give me the URL.

Claude will call `kite_create_room` and `kite_exec` twice, then return:

```
room URL: http://127.0.0.1:8787/rooms/r_abc123xyz789
```

Open that URL — you'll see two command blocks (one for `uname`, one for
`df`), each with full output rendered with ANSI colour preserved.

## Why route shell through kite?

- Every command Claude runs is **observable**: you watch from the browser
  while it works, not after the fact.
- Commands are **replayable**: scrub the timeline to see exactly what
  happened in what order.
- The shell state **persists** across calls. `cd /tmp` followed by `pwd` in
  a later turn still returns `/tmp`.

## Undo

```bash
kite uninstall claude
```

Restores `~/.claude.json` from the backup created during install.
