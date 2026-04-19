package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestSystemStatusRealtime_StreamE2E(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewRealtimeSystemStatusCollector()
	handler := NewSystemStatusRealtimeHandler(collector)

	r := gin.New()
	r.Use(collector.Middleware())
	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/api/v1/boom", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "boom")
	})
	r.POST("/api/v1/admin/system-status/ws-ticket", handler.IssueWSTicket)
	r.GET("/api/v1/admin/system-status/ws", handler.Stream)

	ts := httptest.NewServer(r)
	defer ts.Close()

	respOK, err := http.Get(ts.URL + "/api/v1/ping")
	if err != nil {
		t.Fatalf("ping request failed: %v", err)
	}
	_ = respOK.Body.Close()

	respErr, err := http.Get(ts.URL + "/api/v1/boom")
	if err != nil {
		t.Fatalf("error request failed: %v", err)
	}
	_ = respErr.Body.Close()

	ticket := issueWSTicket(t, ts.URL)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/admin/system-status/ws?ticket=" + url.QueryEscape(ticket)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		msg := ""
		if resp != nil && resp.Body != nil {
			if b, readErr := io.ReadAll(resp.Body); readErr == nil {
				msg = string(b)
			}
			_ = resp.Body.Close()
		}
		t.Fatalf("ws dial failed: %v, response=%s", err, msg)
	}
	defer conn.Close()

	var first RealtimeSystemStatus
	if err := conn.ReadJSON(&first); err != nil {
		t.Fatalf("failed to read first ws snapshot: %v", err)
	}

	if first.GeneratedAtUnixSec <= 0 {
		t.Fatalf("invalid generated_at_unix_sec: %d", first.GeneratedAtUnixSec)
	}
	if first.CPUCores <= 0 {
		t.Fatalf("cpu_cores should be > 0, got %d", first.CPUCores)
	}
	if first.MemoryTotalBytes == 0 {
		t.Fatalf("memory_total_bytes should be > 0, got %d", first.MemoryTotalBytes)
	}
	if first.MemoryUsedBytes > first.MemoryTotalBytes {
		t.Fatalf("memory_used_bytes should be <= memory_total_bytes, got %d > %d", first.MemoryUsedBytes, first.MemoryTotalBytes)
	}
	if first.ActiveConnections <= 0 {
		t.Fatalf("active_connections should include ws client, got %d", first.ActiveConnections)
	}
	if first.APILatencyMS <= 0 {
		t.Fatalf("api_latency_ms should be > 0 after API traffic, got %.6f", first.APILatencyMS)
	}

	_, reusedResp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("reused ticket should not be accepted")
	}
	if reusedResp == nil {
		t.Fatal("expected HTTP response for rejected reused ticket")
	}
	defer reusedResp.Body.Close()
	if reusedResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for reused ticket, got %d", reusedResp.StatusCode)
	}
}

func issueWSTicket(t *testing.T, baseURL string) string {
	t.Helper()

	resp, err := http.Post(baseURL+"/api/v1/admin/system-status/ws-ticket", "application/json", nil)
	if err != nil {
		t.Fatalf("issue ticket request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("issue ticket status=%d body=%s", resp.StatusCode, string(b))
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Ticket string `json:"ticket"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode ticket response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("unexpected response code: %d", payload.Code)
	}
	if strings.TrimSpace(payload.Data.Ticket) == "" {
		t.Fatal("ticket should not be empty")
	}

	return payload.Data.Ticket
}
