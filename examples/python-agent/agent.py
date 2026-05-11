"""Toy agent driving kite over HTTP.

Requires: requests, websocket-client (`pip install requests websocket-client`).
"""

import json
import sys
import threading
import requests
import websocket

BASE = "http://127.0.0.1:8787"


def watch(room_id: str) -> None:
    url = f"ws://127.0.0.1:8787/api/v1/rooms/{room_id}/stream"

    def on_message(_ws, raw: str) -> None:
        msg = json.loads(raw)
        if msg.get("type") == "event":
            ev = msg["event"]
            if ev["type"] == "command.finished":
                p = ev["payload"]
                print(f"  ⎯ finished {p['command_id']} exit={p['exit_code']}")

    def on_open(_ws) -> None:
        print(f"  ⎯ watching {url}")

    ws = websocket.WebSocketApp(url, on_message=on_message, on_open=on_open)
    threading.Thread(target=ws.run_forever, daemon=True).start()


def main() -> int:
    room = requests.post(f"{BASE}/api/v1/rooms", json={"name": "python-agent"}).json()
    print(f"room: {room['id']}")
    print(f"viewer: {BASE}/rooms/{room['id']}")
    watch(room["id"])

    print("type shell commands; empty line to quit")
    while True:
        try:
            cmd = input("> ")
        except (EOFError, KeyboardInterrupt):
            print()
            break
        if not cmd.strip():
            break

        res = requests.post(
            f"{BASE}/api/v1/rooms/{room['id']}/exec",
            json={"cmd": cmd, "timeout_seconds": 30, "source": "python-agent"},
            timeout=60,
        ).json()
        sys.stdout.write(res.get("stdout", ""))
        sys.stdout.flush()
        print(f"[exit {res['exit_code']} in {res['duration_ms']}ms]")

    requests.delete(f"{BASE}/api/v1/rooms/{room['id']}")
    print("room closed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
