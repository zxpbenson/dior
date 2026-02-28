package main

import (
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFullFlow_FileToFile tests a complete flow reading from a file and writing to a file.
// This serves as an integration test.
func TestFullFlow_FileToFile(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-full-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	srcContent := "line1\nline2\nline3\n"
	if err := os.WriteFile(srcFile, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Define dest file
	dstFile := filepath.Join(tmpDir, "dest.txt")

	// Mock command line arguments
	// We can't easily run main() because it calls os.Exit.
	// Instead, we can simulate the logic inside main or refactor main to be testable.
	// For now, let's just verify that we can parse options and run the controller logic manually,
	// effectively testing the integration of components.

	opts := option.NewOptions("dior-test")
	opts.Src = "press"
	opts.SrcFile = srcFile
	opts.SrcSpeed = 100
	opts.SrcScannerBufSizeMb = 1
	opts.Dst = "file"
	opts.DstFile = dstFile
	opts.DstBufSizeByte = 1024
	opts.ChanSize = 100
	opts.LogLevel = "error" // Keep logs quiet

	// Run the flow (similar to main)
	if err := lg.InitDftLgr(opts.LogPrefix, opts.LogLevel); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	controller := component.NewController()
	if err := controller.AddComponents(opts); err != nil {
		t.Fatalf("Failed to add components: %v", err)
	}
	if err := controller.Init(); err != nil {
		t.Fatalf("Failed to init controller: %v", err)
	}

	// Run async in goroutine so we can stop it
	go controller.Start()

	// Wait for some data to be processed
	// Since we can't easily signal completion in this simple test without modifying Controller to accept a done channel for testing,
	// we'll just wait a bit.
	time.Sleep(500 * time.Millisecond)

	// In a real test, we would trigger a signal to stop controller gracefully.
	// For now, this just verifies that the components can be wired up and started without crashing immediately.

	// Check if destination file was created (it might be empty or partial depending on flush)
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Errorf("Destination file was not created")
	}
}