package gocache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
	"github.com/amirhnajafiz/bedrock-api/internal/storage/gocache"
)

func newTestBackend() *gocache.Backend {
	return gocache.NewBackend(time.Minute)
}

func TestBackend_SetAndGet(t *testing.T) {
	b := newTestBackend()

	if err := b.Set("k1", []byte("hello")); err != nil {
		t.Fatalf("Set: unexpected error: %v", err)
	}

	got, err := b.Get("k1")
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}

	if string(got) != "hello" {
		t.Errorf("Get: got %q, want %q", got, "hello")
	}
}

func TestBackend_Get_NotFound(t *testing.T) {
	b := newTestBackend()

	_, err := b.Get("missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Get missing key: got %v, want storage.ErrNotFound", err)
	}
}

func TestBackend_Set_Overwrites(t *testing.T) {
	b := newTestBackend()

	_ = b.Set("k", []byte("first"))
	_ = b.Set("k", []byte("second"))

	got, err := b.Get("k")
	if err != nil {
		t.Fatalf("Get after overwrite: %v", err)
	}

	if string(got) != "second" {
		t.Errorf("Get after overwrite: got %q, want %q", got, "second")
	}
}

func TestBackend_Delete(t *testing.T) {
	b := newTestBackend()

	_ = b.Set("k", []byte("v"))

	if err := b.Delete("k"); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}

	_, err := b.Get("k")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Get after Delete: got %v, want storage.ErrNotFound", err)
	}
}

func TestBackend_Delete_NoOp(t *testing.T) {
	b := newTestBackend()

	// Deleting a non-existent key must not return an error.
	if err := b.Delete("ghost"); err != nil {
		t.Errorf("Delete missing key: unexpected error: %v", err)
	}
}

func TestBackend_List_WithPrefix(t *testing.T) {
	b := newTestBackend()

	_ = b.Set("sessions/a", []byte("sa"))
	_ = b.Set("sessions/b", []byte("sb"))
	_ = b.Set("events/x", []byte("ex"))

	result, err := b.List("sessions/")
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("List sessions/: got %d entries, want 2", len(result))
	}

	for k, v := range result {
		switch k {
		case "sessions/a":
			if string(v) != "sa" {
				t.Errorf("sessions/a: got %q, want %q", v, "sa")
			}
		case "sessions/b":
			if string(v) != "sb" {
				t.Errorf("sessions/b: got %q, want %q", v, "sb")
			}
		default:
			t.Errorf("List returned unexpected key %q", k)
		}
	}
}

func TestBackend_List_EmptyPrefix(t *testing.T) {
	b := newTestBackend()

	_ = b.Set("a", []byte("1"))
	_ = b.Set("b", []byte("2"))
	_ = b.Set("c", []byte("3"))

	result, err := b.List("")
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("List empty prefix: got %d entries, want 3", len(result))
	}
}

func TestBackend_List_NoMatch(t *testing.T) {
	b := newTestBackend()

	_ = b.Set("sessions/a", []byte("v"))

	result, err := b.List("events/")
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("List no-match: got %d entries, want 0", len(result))
	}
}
