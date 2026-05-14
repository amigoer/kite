#!/usr/bin/env bash
# kite — full feature tour
#
# Walks through every surface kite exposes — CLI, HTTP API, WebSocket,
# multiple rooms, replay — using only bash + curl + jq + the kite binary.
#
# Usage:
#   ./full-tour.sh                     # uses ./bin/kite or $KITE
#   KITE=/usr/local/bin/kite ./full-tour.sh
#   PORT=18888 ./full-tour.sh          # custom daemon port
#
# Requirements: bash 4+, curl, jq, the kite binary on $PATH or ./bin/kite.

set -euo pipefail

# ─── pretty-print helpers ──────────────────────────────────────────────────

if [[ -t 1 ]]; then
  BOLD=$'\e[1m' DIM=$'\e[2m' GREEN=$'\e[32m' BLUE=$'\e[34m' YELLOW=$'\e[33m' RESET=$'\e[0m'
else
  BOLD='' DIM='' GREEN='' BLUE='' YELLOW='' RESET=''
fi

step() { echo; echo "${BOLD}${BLUE}▸ $1${RESET}"; }
ok()   { echo "  ${GREEN}✓${RESET} $1"; }
note() { echo "  ${DIM}$1${RESET}"; }
run()  { echo "  ${YELLOW}\$${RESET} $*"; "$@"; }

# ─── locate kite ───────────────────────────────────────────────────────────

KITE="${KITE:-}"
if [[ -z "$KITE" ]]; then
  if [[ -x "./bin/kite" ]]; then
    KITE="./bin/kite"
  elif command -v kite >/dev/null; then
    KITE="$(command -v kite)"
  else
    echo "kite binary not found. Set KITE=/path/to/kite, or run 'make build' first." >&2
    exit 1
  fi
fi

PORT="${PORT:-18999}"
BASE="http://127.0.0.1:$PORT"

if ! command -v jq >/dev/null; then
  echo "jq is required for this demo" >&2
  exit 1
fi

# ─── start a private daemon ────────────────────────────────────────────────

WORKDIR="$(mktemp -d -t kite-demo.XXXXXX)"
trap 'cleanup' EXIT

cleanup() {
  if [[ -n "${DAEMON_PID:-}" ]]; then
    kill "$DAEMON_PID" 2>/dev/null || true
    wait "$DAEMON_PID" 2>/dev/null || true
  fi
  rm -rf "$WORKDIR"
}

step "Starting a private kite daemon on port $PORT"
KITE_HOME="$WORKDIR" "$KITE" serve --port "$PORT" >"$WORKDIR/daemon.log" 2>&1 &
DAEMON_PID=$!

for _ in {1..30}; do
  if curl -fsS "$BASE/healthz" >/dev/null 2>&1; then
    ok "daemon up at $BASE (pid $DAEMON_PID)"
    break
  fi
  sleep 0.1
done
if ! curl -fsS "$BASE/healthz" >/dev/null 2>&1; then
  echo "daemon failed to start; log:" >&2
  cat "$WORKDIR/daemon.log" >&2
  exit 1
fi

# ─── 1. CLI: create / list / show ──────────────────────────────────────────

step "1. Creating a room via the CLI"
run "$KITE" --port "$PORT" room create --name build
BUILD_ID=$("$KITE" --port "$PORT" room list --json | jq -r '.[] | select(.name=="build") | .id')
ok "build room: $BUILD_ID"

run "$KITE" --port "$PORT" room create --name test
TEST_ID=$("$KITE" --port "$PORT" room list --json | jq -r '.[] | select(.name=="test") | .id')
ok "test room: $TEST_ID"

step "2. Listing rooms"
run "$KITE" --port "$PORT" room list

# ─── 3. HTTP exec ──────────────────────────────────────────────────────────

step "3. Running commands via HTTP /exec"
note "Run a few commands in the 'build' room. State persists between calls."

for cmd in "uname -s" "pwd" "ls / | head -3"; do
  echo
  echo "  ${YELLOW}\$${RESET} exec $BUILD_ID -- $cmd"
  curl -fsS -X POST "$BASE/api/v1/rooms/$BUILD_ID/exec" \
    -H 'Content-Type: application/json' \
    -d "$(jq -nc --arg cmd "$cmd" '{cmd:$cmd,timeout_seconds:5}')" \
    | jq -r '"    stdout: " + (.stdout | gsub("\n";"\\n")) + "\n    exit: \(.exit_code)  duration: \(.duration_ms)ms"'
done

# ─── 4. Persistent shell state ─────────────────────────────────────────────

step "4. Persistent shell state inside a room"
note "Set a variable in one exec, read it in another."

"$KITE" --port "$PORT" exec "$BUILD_ID" -- 'export MESSAGE="kite remembers"' >/dev/null
echo
run "$KITE" --port "$PORT" exec "$BUILD_ID" -- 'echo "stored: $MESSAGE"'

# ─── 5. Independent rooms ──────────────────────────────────────────────────

step "5. Rooms are independent"
"$KITE" --port "$PORT" exec "$BUILD_ID" -- 'cd /tmp' >/dev/null
"$KITE" --port "$PORT" exec "$TEST_ID"  -- 'cd /' >/dev/null
echo
echo "  build room pwd:"
"$KITE" --port "$PORT" exec "$BUILD_ID" -- pwd | sed 's/^/    /'
echo "  test room pwd:"
"$KITE" --port "$PORT" exec "$TEST_ID" -- pwd | sed 's/^/    /'

# ─── 6. Exit codes and timeouts ────────────────────────────────────────────

step "6. Exit codes propagate"
echo
echo "  ${YELLOW}\$${RESET} exec false"
set +e
"$KITE" --port "$PORT" exec "$BUILD_ID" -- false
RC=$?
set -e
echo "    -> CLI returned exit $RC (matches the command's exit code)"

step "7. Timeouts interrupt long commands"
echo
echo "  ${YELLOW}\$${RESET} exec 'sleep 10' --timeout 1"
START=$(date +%s)
set +e
"$KITE" --port "$PORT" exec "$BUILD_ID" --timeout 1 -- sleep 10 >/dev/null 2>&1
TIMEOUT_RC=$?
set -e
ELAPSED=$(( $(date +%s) - START ))
ok "interrupted after ${ELAPSED}s (CLI exit $TIMEOUT_RC)"

# ─── 8. Event stream ───────────────────────────────────────────────────────

step "8. Querying the event log"
EVENT_COUNT=$(curl -fsS "$BASE/api/v1/rooms/$BUILD_ID/events" | jq '.events | length')
ok "build room has $EVENT_COUNT events"
echo "  ${DIM}first 3 event types:${RESET}"
curl -fsS "$BASE/api/v1/rooms/$BUILD_ID/events" \
  | jq -r '.events[:3][] | "    " + .type'

step "9. Command summaries"
echo "  ${DIM}commands run in the build room (derived from events):${RESET}"
curl -fsS "$BASE/api/v1/rooms/$BUILD_ID/commands" \
  | jq -r '.commands[] |
      "    " + .cmd + "  " +
      (if .exit_code != null then "(exit \(.exit_code), \(.duration_ms)ms)" else "(running)" end)'

# ─── 10. Replay in the terminal ────────────────────────────────────────────

step "10. CLI replay (instant)"
note "kite replay walks the event log; --no-timing skips the inter-event delay."
echo
"$KITE" --port "$PORT" replay "$BUILD_ID" --no-timing | sed 's/^/    /'

# ─── 11. Web viewer (just print the URL) ───────────────────────────────────

step "11. Web viewer URLs"
echo "    build:  $BASE/rooms/$BUILD_ID"
echo "    test:   $BASE/rooms/$TEST_ID"
echo "    list:   $BASE/rooms"
note "Open one in a browser to see the live command-block dashboard + replay timeline."

# ─── 12. Doctor ─────────────────────────────────────────────────────────────

step "12. kite doctor"
echo
"$KITE" --port "$PORT" doctor | sed 's/^/    /'

# ─── 13. Cleanup ───────────────────────────────────────────────────────────

step "13. Closing rooms"
"$KITE" --port "$PORT" room close "$BUILD_ID"
"$KITE" --port "$PORT" room close "$TEST_ID"
ok "both rooms closed"

echo
echo "${BOLD}${GREEN}✓ tour complete${RESET}"
echo "  daemon log: $WORKDIR/daemon.log"
echo
