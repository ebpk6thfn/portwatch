package portscanner

import (
	"os"
	"testing"
)

func TestParseProcNet_TCP(t *testing.T) {
	// Sample /proc/net/tcp format
	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 0000000000000000 100 0 0 10 0
   1: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 23456 1 0000000000000000 100 0 0 10 0
`
	tmpFile, err := os.CreateTemp("", "proc_net_tcp")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	entries, err := parseProcNet(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseProcNet returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", entries[0].Port)
	}
	if entries[0].LocalAddr != "127.0.0.1" {
		t.Errorf("expected addr 127.0.0.1, got %s", entries[0].LocalAddr)
	}

	if entries[1].Port != 80 {
		t.Errorf("expected port 80, got %d", entries[1].Port)
	}
	if entries[1].LocalAddr != "0.0.0.0" {
		t.Errorf("expected addr 0.0.0.0, got %s", entries[1].LocalAddr)
	}
}

func TestParseProcNet_MissingFile(t *testing.T) {
	_, err := parseProcNet("/nonexistent/path/proc/net/tcp")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParsePort_Valid(t *testing.T) {
	tests := []struct {
		hex      string
		expected uint16
	}{
		{"0050", 80},
		{"1F90", 8080},
		{"01BB", 443},
		{"0016", 22},
	}

	for _, tt := range tests {
		got, err := parsePort(tt.hex)
		if err != nil {
			t.Errorf("parsePort(%q) returned error: %v", tt.hex, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("parsePort(%q) = %d, want %d", tt.hex, got, tt.expected)
		}
	}
}

func TestNewScanner_Scan(t *testing.T) {
	scanner := NewScanner()
	if scanner == nil {
		t.Fatal("NewScanner returned nil")
	}

	// Scan should not error on a real system (or gracefully handle missing /proc)
	entries, err := scanner.Scan()
	if err != nil {
		// On non-Linux systems /proc/net/tcp won't exist; that's acceptable
		t.Logf("Scan returned error (may be expected on non-Linux): %v", err)
		return
	}
	// entries can be empty but should not be nil on success
	if entries == nil {
		t.Error("expected non-nil entries slice on success")
	}
}
