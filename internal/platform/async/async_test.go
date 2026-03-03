package async

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRun_NonBlocking(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	begin := time.Now()
	Run(func(ctx context.Context) error {
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}, "non-blocking")
	elapsed := time.Since(begin)

	if elapsed > 50*time.Millisecond {
		t.Fatalf("Run() blocked caller for %v", elapsed)
	}

	select {
	case <-started:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("async function did not start")
	}

	close(release)
}

func TestRun_PanicRecoveredAndProcessContinues(t *testing.T) {
	Run(func(ctx context.Context) error {
		panic("boom")
	}, "panic-task")

	var executed int32
	done := make(chan struct{})
	Run(func(ctx context.Context) error {
		atomic.StoreInt32(&executed, 1)
		close(done)
		return nil
	}, "follow-up")

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("follow-up async task did not complete; panic may not have been recovered")
	}

	if atomic.LoadInt32(&executed) != 1 {
		t.Fatal("follow-up async task did not execute")
	}
}
