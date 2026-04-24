package portscanner

import (
	"testing"
	"time"
)

func TestTagStore_SetAndGet(t *testing.T) {
	ts := NewTagStore()
	ts.Set("tcp:80", []string{"http", "web"}, 0)
	tags, ok := ts.Get("tcp:80")
	if !ok {
		t.Fatal("expected tags to exist")
	}
	if len(tags) != 2 || tags[0] != "http" || tags[1] != "web" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagStore_MissingKey(t *testing.T) {
	ts := NewTagStore()
	_, ok := ts.Get("tcp:9999")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestTagStore_Delete(t *testing.T) {
	ts := NewTagStore()
	ts.Set("tcp:443", []string{"tls"}, 0)
	ts.Delete("tcp:443")
	_, ok := ts.Get("tcp:443")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestTagStore_TTL_Expired(t *testing.T) {
	ts := NewTagStore()
	ts.Set("udp:53", []string{"dns"}, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	_, ok := ts.Get("udp:53")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestTagStore_TTL_NotExpired(t *testing.T) {
	ts := NewTagStore()
	ts.Set("udp:53", []string{"dns"}, 10*time.Second)
	_, ok := ts.Get("udp:53")
	if !ok {
		t.Fatal("expected entry to still be valid")
	}
}

func TestTagStore_Flush_RemovesExpired(t *testing.T) {
	ts := NewTagStore()
	ts.Set("tcp:80", []string{"web"}, 1*time.Millisecond)
	ts.Set("tcp:443", []string{"tls"}, 10*time.Second)
	time.Sleep(5 * time.Millisecond)
	ts.Flush()
	if ts.Len() != 1 {
		t.Fatalf("expected 1 entry after flush, got %d", ts.Len())
	}
	_, ok := ts.Get("tcp:443")
	if !ok {
		t.Fatal("expected non-expired entry to survive flush")
	}
}

func TestTagStore_Len_CountsNonExpired(t *testing.T) {
	ts := NewTagStore()
	ts.Set("a", []string{"x"}, 0)
	ts.Set("b", []string{"y"}, 0)
	ts.Set("c", []string{"z"}, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if ts.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", ts.Len())
	}
}

func TestTagStore_Set_Overwrites(t *testing.T) {
	ts := NewTagStore()
	ts.Set("tcp:22", []string{"ssh"}, 0)
	ts.Set("tcp:22", []string{"ssh", "admin"}, 0)
	tags, ok := ts.Get("tcp:22")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags after overwrite, got %d", len(tags))
	}
}
