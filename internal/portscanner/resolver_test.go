package portscanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// buildFakeProc creates a minimal fake /proc tree for a single process.
func buildFakeProc(t *testing.T, pid int, comm string, inode uint64) string {
	t.Helper()
	root := t.TempDir()
	pidDir := filepath.Join(root, fmt.Sprintf("%d", pid))
	fdDir := filepath.Join(pidDir, "fd")
	if err := os.MkdirAll(fdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write comm file
	if err := os.WriteFile(filepath.Join(pidDir, "comm"), []byte(comm+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a symlink that mimics a socket fd
	link := filepath.Join(fdDir, "3")
	target := fmt.Sprintf("socket:[%d]", inode)
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestResolver_InodeToProcess_Found(t *testing.T) {
	const pid = 1234
	const comm = "nginx"
	const inode = uint64(99887766)

	root := buildFakeProc(t, pid, comm, inode)
	r := NewResolver(root)

	gotPID, gotName := r.InodeToProcess(inode)
	if gotPID != pid {
		t.Errorf("expected pid %d, got %d", pid, gotPID)
	}
	if gotName != comm {
		t.Errorf("expected name %q, got %q", comm, gotName)
	}
}

func TestResolver_InodeToProcess_NotFound(t *testing.T) {
	const pid = 42
	const comm = "sshd"
	const inode = uint64(11111)
	const wrongInode = uint64(99999)

	root := buildFakeProc(t, pid, comm, inode)
	r := NewResolver(root)

	gotPID, gotName := r.InodeToProcess(wrongInode)
	if gotPID != 0 || gotName != "" {
		t.Errorf("expected empty result, got pid=%d name=%q", gotPID, gotName)
	}
}

func TestResolver_EmptyProcRoot(t *testing.T) {
	r := NewResolver(t.TempDir()) // empty dir — no pid subdirs
	gotPID, gotName := r.InodeToProcess(12345)
	if gotPID != 0 || gotName != "" {
		t.Errorf("expected empty result for empty proc root, got pid=%d name=%q", gotPID, gotName)
	}
}

func TestNewResolver_DefaultProcRoot(t *testing.T) {
	r := NewResolver("")
	if r.procRoot != "/proc" {
		t.Errorf("expected default procRoot /proc, got %q", r.procRoot)
	}
}
