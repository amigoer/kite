// A small build-verifier "agent" that uses kite as its shell backend.
//
// Walks a repo, detects its language, runs build + test inside a single
// kite room, and reports back. Demonstrates the patterns you'd build an
// LLM-driven agent on top of:
//
//   1. create one room — shell state persists across commands
//   2. run a sequence of execs, observing exit codes
//   3. subscribe to the WebSocket stream to print events as they happen
//   4. close the room when done
//
// This file is intentionally self-contained: it talks to the daemon over
// plain net/http, so you can copy it into any other project without
// pulling in kite source.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// ─── tiny kite HTTP client ─────────────────────────────────────────────────

type kite struct {
	base   string
	http   *http.Client
}

type room struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type execReq struct {
	Cmd            string `json:"cmd"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	Source         string `json:"source"`
}

type execRes struct {
	CommandID  string `json:"command_id"`
	Stdout     string `json:"stdout"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

func newKite(base string) *kite {
	if base == "" {
		base = "http://127.0.0.1:8787"
	}
	return &kite{base: base, http: &http.Client{Timeout: 0}}
}

func (k *kite) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, _ := http.NewRequestWithContext(ctx, method, k.base+path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := k.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w (is the daemon running? `kite serve`)", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: HTTP %d: %s", method, path, resp.StatusCode, raw)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (k *kite) createRoom(ctx context.Context, name, cwd string) (*room, error) {
	body := map[string]string{"name": name, "cwd": cwd}
	r := &room{}
	return r, k.do(ctx, "POST", "/api/v1/rooms", body, r)
}

func (k *kite) exec(ctx context.Context, roomID, cmd string, timeout int) (*execRes, error) {
	res := &execRes{}
	err := k.do(ctx, "POST", "/api/v1/rooms/"+roomID+"/exec",
		execReq{Cmd: cmd, TimeoutSeconds: timeout, Source: "go-agent"}, res)
	return res, err
}

func (k *kite) closeRoom(ctx context.Context, id string) error {
	return k.do(ctx, "DELETE", "/api/v1/rooms/"+id, nil, nil)
}

// ─── live event subscriber ─────────────────────────────────────────────────

type wsEvent struct {
	Type  string `json:"type"`
	Event struct {
		ID      int64           `json:"id"`
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	} `json:"event"`
}

// watchEvents subscribes to the room's WebSocket and prints one line per
// command.started / command.finished, plus chunks of stdout.
func watchEvents(ctx context.Context, base, roomID string) {
	u, _ := url.Parse(base)
	u.Scheme = "ws"
	u.Path = "/api/v1/rooms/" + roomID + "/stream"

	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "  ⚠ ws dial failed:", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "agent done")

	for {
		var msg wsEvent
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		if msg.Type != "event" {
			continue
		}
		switch msg.Event.Type {
		case "command.started":
			var p struct {
				Cmd string `json:"cmd"`
			}
			_ = json.Unmarshal(msg.Event.Payload, &p)
			fmt.Printf("    ▸ %s\n", p.Cmd)
		case "command.output":
			var p struct {
				Data string `json:"data"`
			}
			_ = json.Unmarshal(msg.Event.Payload, &p)
			if dec, err := base64.StdEncoding.DecodeString(p.Data); err == nil {
				// indent each line for readability
				for _, line := range strings.SplitAfter(string(dec), "\n") {
					if line == "" {
						continue
					}
					fmt.Printf("      │ %s", line)
				}
				if !strings.HasSuffix(string(dec), "\n") {
					fmt.Println()
				}
			}
		case "command.finished":
			var p struct {
				Exit int   `json:"exit_code"`
				Dur  int64 `json:"duration_ms"`
			}
			_ = json.Unmarshal(msg.Event.Payload, &p)
			sym := "✓"
			if p.Exit != 0 {
				sym = "✗"
			}
			fmt.Printf("    %s exit %d, %dms\n", sym, p.Exit, p.Dur)
		}
	}
}

// ─── the "agent" itself ────────────────────────────────────────────────────

type project struct {
	cwd       string
	language  string
	buildCmd  string
	testCmd   string
}

func detect(cwd string) *project {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(cwd, rel))
		return err == nil
	}
	switch {
	case exists("go.mod"):
		return &project{cwd, "Go", "go build ./...", "go test ./... -count=1"}
	case exists("package.json"):
		return &project{cwd, "Node", "npm install --silent", "npm test --silent"}
	case exists("Cargo.toml"):
		return &project{cwd, "Rust", "cargo build", "cargo test"}
	case exists("pyproject.toml") || exists("setup.py"):
		return &project{cwd, "Python", "python -m pip install -e . -q", "pytest -q"}
	default:
		return nil
	}
}

func main() {
	base := flag.String("base", "http://127.0.0.1:8787", "kite daemon base URL")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go-agent [--base URL] <project-dir>")
		os.Exit(2)
	}
	cwd, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		exit(err)
	}
	if _, err := os.Stat(cwd); err != nil {
		exit(err)
	}

	proj := detect(cwd)
	if proj == nil {
		exit(errors.New("don't know how to build this project (no go.mod / package.json / Cargo.toml / pyproject.toml)"))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	k := newKite(*base)

	r, err := k.createRoom(ctx, "go-agent", cwd)
	if err != nil {
		exit(err)
	}
	fmt.Printf("[+] kite room: %s  →  %s%s\n", r.ID, k.base, r.URL)
	defer func() {
		closeCtx, cc := context.WithTimeout(context.Background(), 3*time.Second)
		defer cc()
		_ = k.closeRoom(closeCtx, r.ID)
		fmt.Println("[+] room closed")
	}()

	// live event watcher in the background
	go watchEvents(ctx, k.base, r.ID)
	time.Sleep(200 * time.Millisecond) // let WS init message arrive first

	fmt.Println("[+] step 1/3: detect language")
	fmt.Printf("    found %s project\n", proj.language)

	fmt.Println("[+] step 2/3: build")
	build, err := k.exec(ctx, r.ID, proj.buildCmd, 120)
	if err != nil {
		exit(err)
	}

	fmt.Println("[+] step 3/3: test")
	test, err := k.exec(ctx, r.ID, proj.testCmd, 300)
	if err != nil {
		exit(err)
	}

	// Give the WS watcher a moment to drain the last events.
	time.Sleep(200 * time.Millisecond)

	fmt.Println("[+] report:")
	status := func(r *execRes, name string) string {
		if r.ExitCode == 0 {
			return fmt.Sprintf("    %-7s pass (%dms)", name+":", r.DurationMs)
		}
		return fmt.Sprintf("    %-7s FAIL (exit %d, %dms)", name+":", r.ExitCode, r.DurationMs)
	}
	fmt.Println(status(build, "BUILD"))
	fmt.Println(status(test, "TEST"))
	fmt.Printf("    See %s%s for full output.\n", k.base, r.URL)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
