package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCache_LoadAndRetrieve(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-test-cache")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	inFile := filepath.Join(tmpDir, "data.txt")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(inFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Cache
	c := NewCache(1) // 1MB buffer

	// Load Data
	if err := c.Load(inFile); err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	// Verify size
	if c.size != 3 {
		t.Errorf("Expected size 3, got %d", c.size)
	}

	// Retrieve Data using Ticket
	ticket := NewTicket()

	// 1st item
	data1 := c.Next(ticket)
	if string(data1) != "line1" {
		t.Errorf("Expected 'line1', got '%s'", string(data1))
	}
	if ticket.index != 1 {
		t.Errorf("Expected ticket index 1, got %d", ticket.index)
	}

	// 2nd item
	data2 := c.Next(ticket)
	if string(data2) != "line2" {
		t.Errorf("Expected 'line2', got '%s'", string(data2))
	}

	// 3rd item
	data3 := c.Next(ticket)
	if string(data3) != "line3" {
		t.Errorf("Expected 'line3', got '%s'", string(data3))
	}

	// 4th item (should loop back to 1st)
	data4 := c.Next(ticket)
	if string(data4) != "line1" {
		t.Errorf("Expected 'line1', got '%s'", string(data4))
	}
	if ticket.index != 1 {
		t.Errorf("Expected ticket index 1 (reset), got %d", ticket.index)
	}
}

func TestCache_LoadEmpty(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-test-cache-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create empty file
	inFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(inFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewCache(1)
	if err := c.Load(inFile); err != nil {
		t.Fatalf("Failed to load empty file: %v", err)
	}

	if c.size != 0 {
		t.Errorf("Expected size 0, got %d", c.size)
	}
}

func BenchmarkCache_Next(b *testing.B) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "dior-bench-cache")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file with some data
	inFile := filepath.Join(tmpDir, "bench_data.txt")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(inFile, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	c := NewCache(1)
	if err := c.Load(inFile); err != nil {
		b.Fatalf("Failed to load cache: %v", err)
	}

	ticket := NewTicket()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = c.Next(ticket)
	}
}