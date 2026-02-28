package source

import (
	"context"
	"dior/internal/lg"
	"dior/option"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func init() {
	lg.InitDftLgr("test", "error")
}

func TestPressSource_Lifecycle(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-test-source")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test input file
	inFile := filepath.Join(tmpDir, "input.txt")
	testData := []byte("line1\nline2\nline3\n")
	if err := os.WriteFile(inFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	opts := &option.Options{
		SrcFile:             inFile,
		SrcSpeed:            10,
		SrcScannerBufSizeMb: 1,
	}

	// Create Source
	srcCmp, err := newPressSource(opts)
	if err != nil {
		t.Fatalf("Failed to create press source: %v", err)
	}
	source, ok := srcCmp.(*PressSource)
	if !ok {
		t.Fatal("Component is not of type *PressSource")
	}

	// Setup Controller (WaitGroup)
	wg := &sync.WaitGroup{}
	source.UnderControl(wg)

	// Init Source with Channel
	dataCh := make(chan []byte, 100)
	if err := source.Init(dataCh); err != nil {
		t.Fatalf("Failed to init press source: %v", err)
	}

	// Start Source
	ctx, cancel := context.WithCancel(context.Background())
	source.Start(ctx)

	// Wait for data generation
	// We expect 3 lines
	expectedLines := 3
	receivedCount := 0
	
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	// Since speed is limited, it might take a bit. But for 3 lines at 10/s, it should be fast.
	// Wait for at least one batch
	time.Sleep(200 * time.Millisecond)

LOOP:
	for {
		select {
		case data := <-dataCh:
			t.Logf("Received: %s", string(data))
			receivedCount++
			if receivedCount >= expectedLines {
				// We might get more because press source loops the file content if needed or just repeats?
				// Looking at press.go logic:
				// writeLoop iterates `cnt <= limit`. `cache.Next(ticket)` retrieves data.
				// It seems `cache` cycles through data.
				break LOOP
			}
		case <-timer.C:
			t.Logf("Timeout waiting for data. Received %d lines", receivedCount)
			if receivedCount == 0 {
				t.Error("Did not receive any data")
			}
			break LOOP
		}
	}

	// Simulate graceful shutdown sequence
	cancel()
	source.Stop() // For sources, Stop() is usually called to flush/close before channel close

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for source to stop")
	}
	
	// Close channel after source has stopped writing
	close(dataCh)
}