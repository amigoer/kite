package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/amigoer/kite/internal/room"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestCreateAndGetRoom(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r := &room.Room{
		ID:        "r_test01",
		Name:      "demo",
		CreatedAt: time.Now(),
		Status:    room.StatusActive,
		Cwd:       "/tmp",
		Shell:     "/bin/bash",
		Metadata:  map[string]string{"foo": "bar"},
	}
	if err := s.CreateRoom(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.GetRoom(ctx, "r_test01")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "demo" || got.Cwd != "/tmp" || got.Status != room.StatusActive {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.Metadata["foo"] != "bar" {
		t.Errorf("metadata lost: %+v", got.Metadata)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetRoom(context.Background(), "r_missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestListRoomsByStatus(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for i, st := range []room.Status{room.StatusActive, room.StatusClosed, room.StatusActive} {
		err := s.CreateRoom(ctx, &room.Room{
			ID:        []string{"r_a", "r_b", "r_c"}[i],
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			Status:    st,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	got, err := s.ListRooms(ctx, room.ListRoomsFilter{Status: room.StatusActive})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 active, got %d", len(got))
	}
}

func TestUpdateRoomStatusClosed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	r := &room.Room{ID: "r_x", CreatedAt: time.Now(), Status: room.StatusActive}
	if err := s.CreateRoom(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.UpdateRoomStatus(ctx, "r_x", room.StatusClosed); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := s.GetRoom(ctx, "r_x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != room.StatusClosed {
		t.Errorf("want closed, got %s", got.Status)
	}
	if got.ClosedAt == nil {
		t.Error("ClosedAt should be set")
	}
}

func TestAppendAndGetEvents(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.CreateRoom(ctx, &room.Room{ID: "r_a", CreatedAt: time.Now(), Status: room.StatusActive}); err != nil {
		t.Fatalf("create: %v", err)
	}

	payload, _ := json.Marshal(room.CommandStartedPayload{CommandID: "c_1", Cmd: "ls"})
	ev := &room.Event{RoomID: "r_a", Type: room.EvtCommandStarted, Payload: payload}
	if err := s.AppendEvent(ctx, ev); err != nil {
		t.Fatalf("append: %v", err)
	}
	if ev.ID == 0 {
		t.Error("expected event ID to be assigned")
	}

	got, err := s.GetEvents(ctx, "r_a", room.GetEventsFilter{})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 || got[0].Type != room.EvtCommandStarted {
		t.Fatalf("unexpected events: %+v", got)
	}
}

func TestGetEventsAfterID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.CreateRoom(ctx, &room.Room{ID: "r_a", CreatedAt: time.Now(), Status: room.StatusActive}); err != nil {
		t.Fatalf("create: %v", err)
	}
	for i := 0; i < 5; i++ {
		if err := s.AppendEvent(ctx, &room.Event{
			RoomID: "r_a", Type: room.EvtCommandOutput, Payload: json.RawMessage(`{}`),
		}); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	got, err := s.GetEvents(ctx, "r_a", room.GetEventsFilter{AfterID: 2, Limit: 10})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
}

func TestSubscribeReceivesEvents(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.CreateRoom(ctx, &room.Room{ID: "r_a", CreatedAt: time.Now(), Status: room.StatusActive}); err != nil {
		t.Fatalf("create: %v", err)
	}

	ch, cancel := s.SubscribeEvents("r_a")
	defer cancel()

	go func() {
		_ = s.AppendEvent(ctx, &room.Event{RoomID: "r_a", Type: room.EvtRoomCreated, Payload: json.RawMessage(`{}`)})
	}()

	select {
	case ev := <-ch:
		if ev.Type != room.EvtRoomCreated {
			t.Errorf("unexpected event type: %s", ev.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive event")
	}
}

func TestSubscribeFiltersByRoom(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for _, id := range []string{"r_a", "r_b"} {
		_ = s.CreateRoom(ctx, &room.Room{ID: id, CreatedAt: time.Now(), Status: room.StatusActive})
	}
	chA, cancelA := s.SubscribeEvents("r_a")
	defer cancelA()

	_ = s.AppendEvent(ctx, &room.Event{RoomID: "r_b", Type: room.EvtCommandStarted, Payload: json.RawMessage(`{}`)})

	select {
	case ev := <-chA:
		t.Fatalf("subscriber should not see other rooms, got %+v", ev)
	case <-time.After(100 * time.Millisecond):
	}
}
