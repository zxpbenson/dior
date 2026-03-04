package component

import (
	"context"
	"dior/internal/lg"
	"sync"
	"testing"
	"time"
)

func init() {
	lg.InitDftLgr("test", "error")
}

func TestAsynchronizer_Work(t *testing.T) {
	wg := &sync.WaitGroup{}
	dataCh := make(chan []byte, 10)
	resultCh := make(chan []byte, 10)

	async := &Asynchronizer{
		Channel: dataCh,
		Output: func(data []byte) error {
			resultCh <- data
			return nil
		},
	}
	async.UnderControl(wg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	async.Start(ctx)

	// Test sending data
	testData := []byte("test-data")
	dataCh <- testData

	select {
	case received := <-resultCh:
		if string(received) != string(testData) {
			t.Errorf("expected %s, got %s", testData, received)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for data")
	}

	// Test cancellation
	cancel()

	// In the new graceful shutdown model, the source stops and then the channel is closed.
	// Asynchronizer relies on channel closure to exit.
	close(dataCh)

	// Wait for goroutine to finish (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for shutdown")
	}
}

func TestAsynchronizer_ChannelClose(t *testing.T) {
	wg := &sync.WaitGroup{}
	dataCh := make(chan []byte, 10)

	async := &Asynchronizer{
		Channel: dataCh,
		Output: func(data []byte) error {
			// Do nothing
			return nil
		},
	}
	async.UnderControl(wg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	async.Start(ctx)

	// Test channel closing
	close(dataCh)

	// Wait for goroutine to finish (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for shutdown after channel close")
	}
}

func BenchmarkAsynchronizer_Work(b *testing.B) {
	lg.InitDftLgr("test", "error")
	wg := &sync.WaitGroup{}
	dataCh := make(chan []byte, 1000)

	async := &Asynchronizer{
		Channel: dataCh,
		Output: func(data []byte) error {
			// No-op for benchmark
			return nil
		},
	}
	async.UnderControl(wg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	async.Start(ctx)

	data := []byte("bench-data")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dataCh <- data
	}

	// Clean up
	cancel()
	go func() {
		wg.Wait()
	}()
}
