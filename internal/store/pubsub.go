package store

import (
	"sync"
	"sync/atomic"

	"github.com/amigoer/kite/internal/room"
)

type subscriber struct {
	id     uint64
	roomID string
	ch     chan *room.Event
}

type bus struct {
	mu      sync.RWMutex
	nextID  uint64
	subs    map[uint64]*subscriber
	closed  atomic.Bool
}

func newBus() *bus {
	return &bus{subs: make(map[uint64]*subscriber)}
}

// subscribe returns a buffered channel that receives events for the given
// room. roomID == "" subscribes to all rooms.
func (b *bus) subscribe(roomID string) (<-chan *room.Event, func()) {
	id := atomic.AddUint64(&b.nextID, 1)
	sub := &subscriber{
		id:     id,
		roomID: roomID,
		ch:     make(chan *room.Event, 128),
	}
	b.mu.Lock()
	b.subs[id] = sub
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		if _, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(sub.ch)
		}
		b.mu.Unlock()
	}
	return sub.ch, cancel
}

func (b *bus) publish(ev *room.Event) {
	if b.closed.Load() {
		return
	}
	b.mu.RLock()
	subs := make([]*subscriber, 0, len(b.subs))
	for _, s := range b.subs {
		if s.roomID == "" || s.roomID == ev.RoomID {
			subs = append(subs, s)
		}
	}
	b.mu.RUnlock()

	for _, s := range subs {
		select {
		case s.ch <- ev:
		default:
			// Slow subscriber: drop. The room state is reconstructable from
			// GetEvents — they can catch up by polling after_id.
		}
	}
}

func (b *bus) closeAll() {
	if !b.closed.CompareAndSwap(false, true) {
		return
	}
	b.mu.Lock()
	for id, s := range b.subs {
		close(s.ch)
		delete(b.subs, id)
	}
	b.mu.Unlock()
}
