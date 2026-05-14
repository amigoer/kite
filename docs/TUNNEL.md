# Reverse tunnel: NAT'd daemons + public hub

Most machines where you want to run a kite daemon don't have a public IP —
your laptop, a home server, a cloud VM you didn't pay extra for. The
machines you want to reach them from often don't either: your phone, a
work laptop on a corporate network, a teammate's browser.

The shape of the problem is:

> Multiple terminal devices behind NAT, plus a small public-IP host I
> control. Everything should route through that host.

`kite hub` is the public-IP host. Daemons dial out to it; browsers
connect to it; it routes between them.

## Three actors

```
   Browser  ──HTTPS──▶  hub.example.com   (kite hub, public IP)
                            │
                            │  yamux multiplexer per daemon
                            ▲
                            │ ① the only inbound connections are
                            │   daemons' outbound WSS upgrades
                            │
              ┌─────────────┼─────────────────┐
              │             │                 │
        laptop (NAT)   home-server (NAT)  work-mac (NAT)
        kite serve     kite serve         kite serve
         --name laptop  --name home-server  --name work-mac
         + their PTYs    + their PTYs        + their PTYs
```

| Component | Role |
| --- | --- |
| **daemon** (`kite serve`) | Runs on the machine that owns the PTYs and SQLite. Dials out to the hub; never accepts inbound traffic from the internet. |
| **hub** (`kite hub`)      | Public-facing relay. Holds long-lived WSS sessions per daemon, routes `/d/<name>/...` to the right one, serves the panel UI. Stateless — restartable freely. |
| **panel** (web bundle)    | The browser SPA, served by the hub. Same code as what a standalone daemon serves, but lives under `/d/<name>/`. |

## Identity model

Each daemon is identified by a **URL-addressable name** plus a **shared
secret**. Both are configured on the hub side via repeatable `--daemon`
flags:

```bash
kite hub --listen :8080 \
  --daemon laptop=$TOKEN_LAPTOP \
  --daemon home-server=$TOKEN_SERVER \
  --daemon work-mac=$TOKEN_WORK
```

Names are case-sensitive, must match `[a-z0-9][a-z0-9-]*`, and appear in
URLs (`/d/<name>/`) so they should be human-friendly. Tokens are random
secrets — `openssl rand -hex 32` is fine.

On each daemon machine:

```bash
kite serve --tunnel-url wss://hub.example.com \
           --name laptop \
           --tunnel-token $TOKEN_LAPTOP
```

If you omit `--name`, kite derives one from `os.Hostname()` by lowercasing
and replacing non-alphanumeric characters with hyphens. So a Mac whose
hostname is `MacBook-Pro.local` registers as `macbook-pro-local` by
default. The daemon logs the resolved name at startup so you can see
what it claimed.

### Why explicit names + tokens

* **Names are routable, tokens are auth.** Names appear in URLs and the
  picker UI; tokens never leave the hub allow-list or the daemon's flag.
* **One token leak doesn't compromise the others.** Each daemon has its
  own token. A breach of `home-server`'s secret can't be used to take
  over `laptop`.
* **Reconnect is automatic and safe.** If a daemon re-presents the same
  (name, valid token) pair, the hub evicts any old session for that
  name — clean wake-from-sleep behavior. A wrong token returns 401 and
  is logged, so a brute-force attempt against the hub is visible.
* **Unknown names are 403, never silent.** A daemon trying to register
  under a name not in the allow-list is rejected with a clear log line,
  not auto-promoted to `name-2`.

## Quick start

### On the hub host (public IP)

```bash
TOK1=$(openssl rand -hex 32)
TOK2=$(openssl rand -hex 32)

# In production: put nginx / Caddy / Cloudflare in front for TLS.
kite hub --listen :8080 \
  --daemon laptop=$TOK1 \
  --daemon home-server=$TOK2
```

### On each daemon machine

```bash
kite serve --tunnel-url wss://hub.example.com \
           --name laptop \
           --tunnel-token $TOK1
```

### From any browser

```
https://hub.example.com/
  → picker page listing your daemons (green dot = live)
  → click "laptop" → /d/laptop/  → the full panel UI, talking to that daemon
```

If only one daemon is connected, the picker auto-redirects you straight
into it. Two or more: you choose.

### From the CLI on any machine

```bash
# Talk to a specific daemon through the hub. The --daemon flag becomes
# a /d/<name>/ path-prefix on every request.
kite --host hub.example.com --port 443 --scheme https \
     --daemon laptop \
     room list

kite --host hub.example.com --port 443 --scheme https \
     --daemon laptop \
     attach r_xxx
```

Local-only usage is unchanged: `kite room list` still hits
`http://127.0.0.1:8787` like before.

## What goes over the wire

The hub exposes four kinds of routes:

| Route | Terminates | Notes |
| --- | --- | --- |
| `/_tunnel/connect` | hub | Daemon's WSS upgrade. Daemon presents `X-Kite-Daemon-Name` + `Authorization: Bearer …`. Validated against the `--daemon` allow-list. |
| `/_tunnel/status`  | hub | JSON snapshot of configured daemons + their live status. No auth (no secrets in the response). |
| `/`                | hub | Picker page (HTML+CSS only, no SPA loaded). 0 daemons → empty state. 1 live → 302 to it. ≥2 → picker. |
| `/d/<name>/…`      | proxied | Every other path. The hub strips `/d/<name>` and forwards the rewritten request over a fresh yamux stream to the named daemon. WebSocket upgrades pass through bidirectionally. |
| anything else      | hub | Static asset (favicon, SPA JS/CSS) served from the hub's own embedded bundle. |

The web bundle reads `window.location.pathname` at startup and prefixes
every API/WS URL it constructs with `/d/<name>`. Switching between
daemons is just a navigation back to the picker.

## Keepalive, reconnect, NAT idleness

* **WS-level keepalive**: yamux pings the peer every 15 s
  (`tunnel.YamuxKeepaliveInterval`). NAT routers typically idle-evict
  after 30–120 s; this stays well under.
* **Reconnect**: when the WSS dies the daemon retries with capped
  exponential backoff — 1 s, 2 s, 4 s … 60 s, with ±20 % jitter.
* **Eviction**: a re-registering daemon evicts the previous session for
  its name. In-flight browser requests get a 502 for the few seconds the
  hub has no session, then recover.

## Security model — read before deploying

**This is still MVP.** What the hub authenticates is *daemons*, not
*users*:

* Anyone who can load `https://hub.example.com/d/laptop/` can drive that
  daemon. The hub does **no** browser-side auth.
* The shared `--daemon name=token` pair only protects the daemon → hub
  handshake.
* **Do not** publish a hub URL without an auth layer in front of it.
  Reasonable options:
  * Caddy / nginx with basic auth or mTLS in front of the hub
  * Cloudflare Access / Tailscale Funnel with identity-aware policies
  * VPN — hub binds only to a Tailscale / WireGuard interface

Future work: real browser-side accounts and per-user daemon visibility
(Phase 3). If you publish a hub by accident, **rotate every
`--daemon` token and restart the hub**.

## Local-only dev recipe

To exercise multi-daemon routing without any cloud setup:

```bash
# Terminal 1 — hub on :9090
T1=devsecret-laptop
T2=devsecret-server
./bin/kite hub --listen :9090 \
  --daemon laptop=$T1 \
  --daemon dev-server=$T2 -v

# Terminal 2 — first daemon (uses the default $HOME/.kite data dir)
./bin/kite serve --tunnel-url ws://127.0.0.1:9090 \
  --name laptop --tunnel-token $T1 -v

# Terminal 3 — second daemon (separate data dir + port to avoid the
# single-daemon PID lock on the same machine)
KITE_HOME=/tmp/kite-second-data ./bin/kite serve \
  --port 8788 \
  --tunnel-url ws://127.0.0.1:9090 \
  --name dev-server --tunnel-token $T2 -v

# Terminal 4 — exercise
curl http://127.0.0.1:9090/_tunnel/status
curl -X POST http://127.0.0.1:9090/d/laptop/api/v1/rooms -d '{}'
curl -X POST http://127.0.0.1:9090/d/dev-server/api/v1/rooms -d '{}'
open http://127.0.0.1:9090/
```

The KITE_HOME trick is only needed when two daemons run on the same
machine. Across real machines each one naturally has its own data dir.
