package room

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestArbiterGrantsImmediatelyWhenIdle(t *testing.T) {
	a := newWriteArbiter(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	release, err := a.Claim(ctx, WriteHolder{ID: "a", Kind: "exec"})
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if h := a.Holder(); h == nil || h.ID != "a" {
		t.Fatalf("holder = %v, want a", h)
	}
	release()
	if h := a.Holder(); h != nil {
		t.Fatalf("expected idle holder, got %v", h)
	}
}

func TestArbiterQueuesSecondClaimer(t *testing.T) {
	a := newWriteArbiter(nil)
	r1, err := a.Claim(context.Background(), WriteHolder{ID: "a"})
	if err != nil {
		t.Fatalf("claim a: %v", err)
	}

	got := make(chan struct{})
	go func() {
		r2, err := a.Claim(context.Background(), WriteHolder{ID: "b"})
		if err != nil {
			t.Errorf("claim b: %v", err)
			return
		}
		close(got)
		r2()
	}()

	// b should still be blocked.
	select {
	case <-got:
		t.Fatal("b granted before a released")
	case <-time.After(20 * time.Millisecond):
	}

	r1()
	select {
	case <-got:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("b not granted after a released")
	}
}

func TestArbiterContextCancelLeavesNextWaiterUnaffected(t *testing.T) {
	a := newWriteArbiter(nil)
	r1, _ := a.Claim(context.Background(), WriteHolder{ID: "a"})

	cancelCtx, cancelB := context.WithCancel(context.Background())
	wgB := make(chan error, 1)
	go func() {
		_, err := a.Claim(cancelCtx, WriteHolder{ID: "b"})
		wgB <- err
	}()
	time.Sleep(20 * time.Millisecond) // let b queue
	cancelB()
	if err := <-wgB; err == nil {
		t.Error("expected ctx error for b, got nil")
	}

	gotC := make(chan struct{})
	go func() {
		r, err := a.Claim(context.Background(), WriteHolder{ID: "c"})
		if err != nil {
			t.Errorf("claim c: %v", err)
			return
		}
		close(gotC)
		r()
	}()
	time.Sleep(20 * time.Millisecond) // let c queue
	r1()
	select {
	case <-gotC:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("c not granted after a released and b cancelled")
	}
}

func TestArbiterCloseDrainsWaiters(t *testing.T) {
	a := newWriteArbiter(nil)
	_, _ = a.Claim(context.Background(), WriteHolder{ID: "a"})

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := a.Claim(context.Background(), WriteHolder{ID: "x"})
			errs <- err
		}()
	}
	time.Sleep(20 * time.Millisecond)
	a.Close()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err == nil {
			t.Error("expected error after close, got nil")
		}
	}
}

func TestArbiterNotifiesOnChange(t *testing.T) {
	var (
		mu     sync.Mutex
		events []string
	)
	a := newWriteArbiter(func(h *WriteHolder) {
		mu.Lock()
		defer mu.Unlock()
		if h == nil {
			events = append(events, "idle")
		} else {
			events = append(events, h.ID)
		}
	})

	r, err := a.Claim(context.Background(), WriteHolder{ID: "alpha"})
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	r()
	// onChange is invoked from a goroutine — wait briefly for both events.
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(events)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(events) < 2 || events[0] != "alpha" || events[1] != "idle" {
		t.Errorf("events = %v, want [alpha, idle]", events)
	}
}
