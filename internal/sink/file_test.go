package sink

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

func TestFileSink_Lifecycle(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-test-sink")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "output.txt")
	opts := &option.Options{
		DstFile:        outFile,
		DstBufSizeByte: 1024,
	}

	// Create Sink
	sinkCmp, err := newFileSink("file-sink", opts)
	if err != nil {
		t.Fatalf("Failed to create file sink: %v", err)
	}
	sink, ok := sinkCmp.(*fileSink)
	if !ok {
		t.Fatal("Component is not of type *fileSink")
	}

	// Setup Controller (WaitGroup)
	wg := &sync.WaitGroup{}
	sink.UnderControl(wg)

	// Init Sink with Channel
	dataCh := make(chan []byte, 10)
	if err := sink.Init(dataCh); err != nil {
		t.Fatalf("Failed to init file sink: %v", err)
	}

	// Start Sink
	ctx, cancel := context.WithCancel(context.Background())
	sink.Start(ctx)

	// Send Data
	testData := []byte("hello world")
	dataCh <- testData

	// Wait for processing (since it's async writing)
	// In a real scenario, we might want to ensure flush happens.
	// We'll rely on Stop() to flush.
	time.Sleep(100 * time.Millisecond)

	// Simulate graceful shutdown sequence:
	// 1. Cancel context
	cancel()

	// 2. Close channel (this tells Asynchronizer to finish up and exit its loop)
	close(dataCh)

	// 3. Wait for Asynchronizer goroutine to finish reading remaining data
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for sink to stop")
	}

	// 4. Explicitly call Stop to flush buffered writer and close file
	// (Must be done AFTER wg.Wait() to avoid Data Race)
	sink.Stop()

	// Verify File Content
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := string(testData) + "\n"
	if string(content) != expected {
		t.Errorf("File content mismatch. Expected %q, got %q", expected, string(content))
	}
}
