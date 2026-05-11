# python-agent example

A minimal AI-agent-style loop in Python:

1. create a kite room,
2. ask the user for a command,
3. run it via kite,
4. echo stdout + exit code,
5. subscribe to the WebSocket stream so a human watching the browser sees
   every command as it happens.

## Run

```bash
pip install requests websocket-client
python agent.py
```

Then open `http://127.0.0.1:8787/rooms/<the id printed>` to watch live.

## Files

- `agent.py` — the loop
