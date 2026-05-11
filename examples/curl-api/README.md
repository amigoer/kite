# curl example

Minimal shell session walkthrough using only curl.

```bash
# Start the daemon in another terminal:
kite serve

# Create a room
ROOM=$(curl -s -X POST http://127.0.0.1:8787/api/v1/rooms \
  -H 'Content-Type: application/json' \
  -d '{"name":"curl-demo"}' | jq -r .id)
echo "room: $ROOM"

# Run a command
curl -s -X POST "http://127.0.0.1:8787/api/v1/rooms/$ROOM/exec" \
  -H 'Content-Type: application/json' \
  -d '{"cmd":"ls -la /tmp | head -5","timeout_seconds":10}' | jq

# Pull the event log
curl -s "http://127.0.0.1:8787/api/v1/rooms/$ROOM/events" | jq '.events | length'

# Open the live viewer
echo "view at http://127.0.0.1:8787/rooms/$ROOM"

# Close
curl -s -X DELETE "http://127.0.0.1:8787/api/v1/rooms/$ROOM"
```

Equivalent with the kite CLI:

```bash
kite room create --name curl-demo
kite room list
kite exec <id> -- ls -la /tmp
```
