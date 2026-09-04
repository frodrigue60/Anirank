package og

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTryGenerateRespectsConcurrencyLimit(t *testing.T) {
	g := &Generator{sem: newSemaphore(1)}

	started := make(chan struct{})
	release := make(chan struct{})
	var secondAcquired atomic.Bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ok, err := g.TryGenerate(func() error {
			close(started)
			<-release
			return nil
		})
		if !ok || err != nil {
			t.Errorf("first acquire failed: ok=%v err=%v", ok, err)
		}
	}()

	<-started
	ok, err := g.TryGenerate(func() error {
		secondAcquired.Store(true)
		return nil
	})
	if err != nil {
		t.Fatalf("second TryGenerate err: %v", err)
	}
	if ok {
		t.Fatal("expected second TryGenerate to be rejected while first holds the slot")
	}
	if secondAcquired.Load() {
		t.Fatal("fn must not run when slot unavailable")
	}

	close(release)
	wg.Wait()

	ok, err = g.TryGenerate(func() error { return nil })
	if !ok || err != nil {
		t.Fatalf("expected slot free after release: ok=%v err=%v", ok, err)
	}
}

func TestTryGenerateAllowsUpToMax(t *testing.T) {
	g := &Generator{sem: newSemaphore(2)}
	hold := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := g.TryGenerate(func() error {
				<-hold
				return nil
			})
			if !ok || err != nil {
				t.Errorf("expected acquire within max: ok=%v err=%v", ok, err)
			}
		}()
	}

	// Give goroutines time to acquire both slots.
	time.Sleep(20 * time.Millisecond)
	ok, err := g.TryGenerate(func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("third concurrent generate should be rejected")
	}
	close(hold)
	wg.Wait()
}

func TestNewSemaphoreDefault(t *testing.T) {
	sem := newSemaphore(0)
	if cap(sem) != DefaultMaxConcurrent {
		t.Fatalf("expected default cap %d, got %d", DefaultMaxConcurrent, cap(sem))
	}
}
