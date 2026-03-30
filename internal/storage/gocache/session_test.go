package gocache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
	"github.com/amirhnajafiz/bedrock-api/internal/storage/gocache"
)

func newTestSessionStore() *gocache.SessionStore {
	return gocache.NewSessionStore(gocache.NewBackend(time.Minute))
}

func TestSessionStore_SaveAndGet(t *testing.T) {
	s := newTestSessionStore()

	if err := s.SaveSession("s1", []byte(`{"user":"alice"}`)); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	got, err := s.GetSession("s1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if string(got) != `{"user":"alice"}` {
		t.Errorf("GetSession: got %q", got)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	s := newTestSessionStore()

	_, err := s.GetSession("nope")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetSession missing: got %v, want storage.ErrNotFound", err)
	}
}

func TestSessionStore_Save_Overwrite(t *testing.T) {
	s := newTestSessionStore()

	_ = s.SaveSession("s1", []byte("v1"))
	_ = s.SaveSession("s1", []byte("v2"))

	got, err := s.GetSession("s1")
	if err != nil {
		t.Fatalf("GetSession after overwrite: %v", err)
	}

	if string(got) != "v2" {
		t.Errorf("GetSession after overwrite: got %q, want %q", got, "v2")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	s := newTestSessionStore()

	_ = s.SaveSession("s1", []byte("v"))

	if err := s.DeleteSession("s1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	_, err := s.GetSession("s1")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetSession after delete: got %v, want storage.ErrNotFound", err)
	}
}

func TestSessionStore_Delete_NoOp(t *testing.T) {
	s := newTestSessionStore()

	if err := s.DeleteSession("ghost"); err != nil {
		t.Errorf("DeleteSession missing: unexpected error: %v", err)
	}
}

func TestSessionStore_ListSessions(t *testing.T) {
	s := newTestSessionStore()

	_ = s.SaveSession("s1", []byte("a"))
	_ = s.SaveSession("s2", []byte("b"))
	_ = s.SaveSession("s3", []byte("c"))

	all, err := s.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("ListSessions: got %d entries, want 3", len(all))
	}

	for _, id := range []string{"s1", "s2", "s3"} {
		if _, ok := all[id]; !ok {
			t.Errorf("ListSessions: missing id %q", id)
		}
	}
}

func TestSessionStore_ListSessions_Empty(t *testing.T) {
	s := newTestSessionStore()

	all, err := s.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions empty: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("ListSessions empty: got %d entries, want 0", len(all))
	}
}

func TestSessionStore_ListSessions_IsolatedFromEvents(t *testing.T) {
	// Both stores share the same backend to verify prefix isolation.
	backend := gocache.NewBackend(time.Minute)
	sessions := gocache.NewSessionStore(backend)
	events := gocache.NewEventStore(backend)

	_ = sessions.SaveSession("s1", []byte("session"))
	_ = events.SaveEvent("e1", []byte("event"))

	all, err := sessions.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("ListSessions should not include event keys; got %d entries", len(all))
	}
}