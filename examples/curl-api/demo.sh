#!/usr/bin/env bash
# Runnable smoke test of the kite HTTP API.
# Requires: bash, curl, jq, and a kite daemon listening on 127.0.0.1:8787.

set -euo pipefail

BASE=${BASE:-http://127.0.0.1:8787}

if ! curl -fsS "$BASE/healthz" >/dev/null; then
  echo "kite daemon is not reachable at $BASE. Start it with: kite serve" >&2
  exit 1
fi

ROOM=$(curl -fsS -X POST "$BASE/api/v1/rooms" \
  -H 'Content-Type: application/json' \
  -d '{"name":"curl-demo"}' | jq -r .id)
echo "[+] created room $ROOM"

curl -fsS -X POST "$BASE/api/v1/rooms/$ROOM/exec" \
  -H 'Content-Type: application/json' \
  -d '{"cmd":"echo hello && uname -a","timeout_seconds":5}' | jq

EVENTS=$(curl -fsS "$BASE/api/v1/rooms/$ROOM/events" | jq '.events | length')
echo "[+] $EVENTS events in the log"

echo "[+] open $BASE/rooms/$ROOM in a browser to watch live"

curl -fsS -X DELETE "$BASE/api/v1/rooms/$ROOM" >/dev/null
echo "[+] closed room"
