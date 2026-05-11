// Package store persists rooms and events in SQLite.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/amigoer/kite/internal/room"
)

//go:embed schema.sql
var schemaSQL string

// ErrNotFound is returned when a room does not exist.
var ErrNotFound = errors.New("room not found")

// Store is the SQLite-backed persistence layer.
type Store struct {
	db     *sql.DB
	bus    *bus
	dbPath string
}

// Open opens (and migrates) the store at the given filesystem path. Pass
// ":memory:" for an ephemeral in-memory database (test only).
func Open(ctx context.Context, path string) (*Store, error) {
	dsn := path
	if path != ":memory:" {
		dsn = fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)&_pragma=busy_timeout(5000)", path)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &Store{db: db, bus: newBus(), dbPath: path}, nil
}

// Close releases the database handle and tears down all subscriptions.
func (s *Store) Close() error {
	s.bus.closeAll()
	return s.db.Close()
}

// CreateRoom persists a new room. The CreatedAt and Status fields must be set
// by the caller.
func (s *Store) CreateRoom(ctx context.Context, r *room.Room) error {
	meta, err := marshalMetadata(r.Metadata)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO rooms(id, name, created_at, status, cwd, shell, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID, nullString(r.Name), r.CreatedAt.UnixMilli(), string(r.Status),
		nullString(r.Cwd), nullString(r.Shell), meta,
	)
	if err != nil {
		return fmt.Errorf("insert room: %w", err)
	}
	return nil
}

// GetRoom fetches a room by ID. Returns ErrNotFound if missing.
func (s *Store) GetRoom(ctx context.Context, id string) (*room.Room, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, closed_at, status, cwd, shell, metadata
		 FROM rooms WHERE id = ?`, id)
	return scanRoom(row)
}

// ListRooms returns rooms ordered by created_at descending.
func (s *Store) ListRooms(ctx context.Context, filter room.ListRoomsFilter) ([]*room.Room, error) {
	q := `SELECT id, name, created_at, closed_at, status, cwd, shell, metadata FROM rooms`
	args := []any{}
	if filter.Status != "" {
		q += ` WHERE status = ?`
		args = append(args, string(filter.Status))
	}
	q += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query rooms: %w", err)
	}
	defer rows.Close()

	var out []*room.Room
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateRoomStatus updates a room's status, setting closed_at if status=closed.
func (s *Store) UpdateRoomStatus(ctx context.Context, id string, status room.Status) error {
	if status == room.StatusClosed {
		_, err := s.db.ExecContext(ctx,
			`UPDATE rooms SET status = ?, closed_at = ? WHERE id = ?`,
			string(status), time.Now().UnixMilli(), id)
		if err != nil {
			return fmt.Errorf("update room: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE rooms SET status = ? WHERE id = ?`, string(status), id)
	if err != nil {
		return fmt.Errorf("update room: %w", err)
	}
	return nil
}

// UpdateRoomCwd records the latest cwd for the room (best-effort hint).
func (s *Store) UpdateRoomCwd(ctx context.Context, id, cwd string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE rooms SET cwd = ? WHERE id = ?`, cwd, id)
	if err != nil {
		return fmt.Errorf("update cwd: %w", err)
	}
	return nil
}

// AppendEvent persists an event for the given room, sets its ID and
// timestamp, and broadcasts it to subscribers.
func (s *Store) AppendEvent(ctx context.Context, ev *room.Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO events(room_id, timestamp, type, payload) VALUES (?, ?, ?, ?)`,
		ev.RoomID, ev.Timestamp.UnixMilli(), string(ev.Type), []byte(ev.Payload),
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last id: %w", err)
	}
	ev.ID = id
	s.bus.publish(ev)
	return nil
}

// GetEvents returns events for a room with optional filters, ordered by id asc.
func (s *Store) GetEvents(ctx context.Context, roomID string, filter room.GetEventsFilter) ([]*room.Event, error) {
	q := `SELECT id, room_id, timestamp, type, payload FROM events WHERE room_id = ?`
	args := []any{roomID}
	if filter.AfterID > 0 {
		q += ` AND id > ?`
		args = append(args, filter.AfterID)
	}
	if filter.Type != "" {
		q += ` AND type = ?`
		args = append(args, string(filter.Type))
	}
	q += ` ORDER BY id ASC`
	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var out []*room.Event
	for rows.Next() {
		ev := &room.Event{}
		var ts int64
		var typ string
		var payload []byte
		if err := rows.Scan(&ev.ID, &ev.RoomID, &ts, &typ, &payload); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		ev.Timestamp = time.UnixMilli(ts)
		ev.Type = room.EventType(typ)
		ev.Payload = json.RawMessage(payload)
		out = append(out, ev)
	}
	return out, rows.Err()
}

// SubscribeEvents returns a channel that receives events for the given room.
// Pass "" for roomID to subscribe to all rooms. The caller must call cancel
// to free resources.
func (s *Store) SubscribeEvents(roomID string) (<-chan *room.Event, func()) {
	return s.bus.subscribe(roomID)
}

// Path returns the on-disk path for the database (":memory:" if in-memory).
func (s *Store) Path() string { return s.dbPath }

// --- internal helpers ---------------------------------------------------

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func marshalMetadata(m map[string]string) (any, error) {
	if len(m) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}
	return string(b), nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRoom(row rowScanner) (*room.Room, error) {
	var (
		r           room.Room
		name        sql.NullString
		closedAt    sql.NullInt64
		cwd, shell  sql.NullString
		metadata    sql.NullString
		createdAt   int64
		status      string
	)
	if err := row.Scan(&r.ID, &name, &createdAt, &closedAt, &status, &cwd, &shell, &metadata); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan room: %w", err)
	}
	r.Name = name.String
	r.CreatedAt = time.UnixMilli(createdAt)
	if closedAt.Valid {
		t := time.UnixMilli(closedAt.Int64)
		r.ClosedAt = &t
	}
	r.Status = room.Status(status)
	r.Cwd = cwd.String
	r.Shell = shell.String
	if metadata.Valid && metadata.String != "" {
		m := map[string]string{}
		if err := json.Unmarshal([]byte(metadata.String), &m); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
		r.Metadata = m
	}
	return &r, nil
}
