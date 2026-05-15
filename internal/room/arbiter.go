package room

import (
	"context"
	"errors"
	"sync"
)

// ErrArbiterClosed is returned by writeArbiter operations after Close has been
// called on the parent room session.
var ErrArbiterClosed = errors.New("write arbiter closed")

// WriteHolder describes the client currently holding (or queued for) the
// write claim on a room.
type WriteHolder struct {
	// ID is a daemon-unique handle for the holder; it appears in events so
	// other clients can identify who's typing.
	ID string
	// Kind is one of "exec", "attach", or "web".
	Kind string
	// Label is a short human-readable string (e.g. the cmd line for exec,
	// the source name for attach).
	Label string
}

// writeArbiter serialises stdin access to a room's PTY. At most one holder
// is "active" at a time; everyone else queues FIFO. The arbiter is purely
// cooperative — it grants claims and trusts holders to release in a
// reasonable time. Phase 2 does not implement preemption.
type writeArbiter struct {
	mu        sync.Mutex
	closed    chan struct{}
	closeOnce sync.Once
	current   *claim
	queue     []*claim
	// onChange fires whenever the current holder changes. Invocations are
	// serialised through notifyCh so write.claimed / write.released events
	// are appended in the same order they occurred.
	onChange   func(holder *WriteHolder)
	notifyCh   chan *WriteHolder
	notifyDone chan struct{}
}

type claim struct {
	holder    WriteHolder
	granted   chan struct{} // closed when this claim becomes current
	abandoned bool          // set under arbiter.mu when the waiter gave up
}

func newWriteArbiter(onChange func(*WriteHolder)) *writeArbiter {
	a := &writeArbiter{
		closed:     make(chan struct{}),
		onChange:   onChange,
		notifyCh:   make(chan *WriteHolder, 64),
		notifyDone: make(chan struct{}),
	}
	go a.notifier()
	return a
}

// notifier consumes events from notifyCh and invokes onChange in order.
// Runs until notifyCh is closed (by Close).
func (a *writeArbiter) notifier() {
	defer close(a.notifyDone)
	for h := range a.notifyCh {
		if a.onChange != nil {
			a.onChange(h)
		}
	}
}

// Claim blocks until the caller becomes the active holder, ctx is cancelled,
// or the arbiter is closed. On success it returns a release function that
// hands the claim to the next waiter (or marks the room idle if none).
func (a *writeArbiter) Claim(ctx context.Context, holder WriteHolder) (release func(), err error) {
	c := &claim{
		holder:  holder,
		granted: make(chan struct{}),
	}

	a.mu.Lock()
	select {
	case <-a.closed:
		a.mu.Unlock()
		return nil, ErrArbiterClosed
	default:
	}
	if a.current == nil {
		a.current = c
		close(c.granted)
		a.notifyLocked(&c.holder)
	} else {
		a.queue = append(a.queue, c)
	}
	a.mu.Unlock()

	select {
	case <-c.granted:
		return a.releaseFunc(c), nil
	case <-a.closed:
		return nil, ErrArbiterClosed
	case <-ctx.Done():
		a.abandon(c)
		return nil, ctx.Err()
	}
}

// Holder returns the active holder, or nil when the room is idle.
func (a *writeArbiter) Holder() *WriteHolder {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.current == nil {
		return nil
	}
	h := a.current.holder
	return &h
}

// Close drains all pending waiters with ErrArbiterClosed and prevents future
// claims. Blocks until the notifier goroutine has flushed any pending
// callbacks. Safe to call multiple times.
func (a *writeArbiter) Close() {
	a.closeOnce.Do(func() {
		close(a.closed)
		a.mu.Lock()
		a.current = nil
		a.queue = nil
		// Close notifyCh under the lock so any in-flight notifyLocked
		// (always called under a.mu) finishes its send first.
		close(a.notifyCh)
		a.mu.Unlock()
		<-a.notifyDone
	})
}

func (a *writeArbiter) releaseFunc(c *claim) func() {
	var once sync.Once
	return func() {
		once.Do(func() { a.handoff(c) })
	}
}

func (a *writeArbiter) handoff(c *claim) {
	a.mu.Lock()
	if a.current != c {
		// Close already cleared things; nothing to do.
		a.mu.Unlock()
		return
	}
	a.current = nil
	for len(a.queue) > 0 {
		next := a.queue[0]
		a.queue = a.queue[1:]
		if next.abandoned {
			continue
		}
		a.current = next
		close(next.granted)
		a.notifyLocked(&next.holder)
		a.mu.Unlock()
		return
	}
	a.notifyLocked(nil)
	a.mu.Unlock()
}

func (a *writeArbiter) abandon(c *claim) {
	a.mu.Lock()
	if a.current == c {
		a.mu.Unlock()
		a.handoff(c)
		return
	}
	c.abandoned = true
	for i, q := range a.queue {
		if q == c {
			a.queue = append(a.queue[:i], a.queue[i+1:]...)
			break
		}
	}
	a.mu.Unlock()
}

func (a *writeArbiter) notifyLocked(holder *WriteHolder) {
	if a.onChange == nil {
		return
	}
	// Blocking send onto the buffered channel under a.mu. The dedicated
	// notifier goroutine drains it and invokes onChange in FIFO order, so
	// write.* events stay in claim-history order. Buffer is 64; the only
	// way to block here is if onChange itself is slow, which is acceptable.
	a.notifyCh <- holder
}
