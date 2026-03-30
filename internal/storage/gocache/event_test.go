package gocache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
	"github.com/amirhnajafiz/bedrock-api/internal/storage/gocache"
)

func newTestEventStore() *gocache.EventStore {
	return gocache.NewEventStore(gocache.NewBackend(time.Minute))
}

func TestEventStore_SaveAndGet(t *testing.T) {
	e := newTestEventStore()

	if err := e.SaveEvent("e1", []byte(`{"type":"login"}`)); err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}

	got, err := e.GetEvent("e1")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}

	if string(got) != `{"type":"login"}` {
		t.Errorf("GetEvent: got %q", got)
	}
}

func TestEventStore_Get_NotFound(t *testing.T) {
	e := newTestEventStore()

	_, err := e.GetEvent("missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetEvent missing: got %v, want storage.ErrNotFound", err)
	}
}

func TestEventStore_SaveSetsStatusPending(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("data"))

	status, err := e.GetStatus("e1")
	if err != nil {
		t.Fatalf("GetStatus after save: %v", err)
	}

	if status != storage.EventStatusPending {
		t.Errorf("initial status: got %q, want %q", status, storage.EventStatusPending)
	}
}

func TestEventStore_MarkProcessed(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("data"))

	if err := e.MarkProcessed("e1"); err != nil {
		t.Fatalf("MarkProcessed: %v", err)
	}

	status, err := e.GetStatus("e1")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status != storage.EventStatusProcessed {
		t.Errorf("status after MarkProcessed: got %q, want %q", status, storage.EventStatusProcessed)
	}
}

func TestEventStore_MarkPending(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("data"))
	_ = e.MarkProcessed("e1")

	if err := e.MarkPending("e1"); err != nil {
		t.Fatalf("MarkPending: %v", err)
	}

	status, err := e.GetStatus("e1")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status != storage.EventStatusPending {
		t.Errorf("status after MarkPending: got %q, want %q", status, storage.EventStatusPending)
	}
}

func TestEventStore_MarkProcessed_NotFound(t *testing.T) {
	e := newTestEventStore()

	err := e.MarkProcessed("ghost")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("MarkProcessed missing: got %v, want storage.ErrNotFound", err)
	}
}

func TestEventStore_MarkPending_NotFound(t *testing.T) {
	e := newTestEventStore()

	err := e.MarkPending("ghost")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("MarkPending missing: got %v, want storage.ErrNotFound", err)
	}
}

func TestEventStore_GetStatus_NotFound(t *testing.T) {
	e := newTestEventStore()

	_, err := e.GetStatus("ghost")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetStatus missing: got %v, want storage.ErrNotFound", err)
	}
}

func TestEventStore_Save_ResetsStatus(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("v1"))
	_ = e.MarkProcessed("e1")

	// Overwrite – status must reset to pending.
	_ = e.SaveEvent("e1", []byte("v2"))

	status, err := e.GetStatus("e1")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status != storage.EventStatusPending {
		t.Errorf("status after re-save: got %q, want %q", status, storage.EventStatusPending)
	}
}

func TestEventStore_Delete(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("data"))

	if err := e.DeleteEvent("e1"); err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}

	_, err := e.GetEvent("e1")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetEvent after delete: got %v, want storage.ErrNotFound", err)
	}

	// Status key must also be removed.
	_, err = e.GetStatus("e1")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetStatus after delete: got %v, want storage.ErrNotFound", err)
	}
}

func TestEventStore_Delete_NoOp(t *testing.T) {
	e := newTestEventStore()

	if err := e.DeleteEvent("ghost"); err != nil {
		t.Errorf("DeleteEvent missing: unexpected error: %v", err)
	}
}

func TestEventStore_ListEvents(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("a"))
	_ = e.SaveEvent("e2", []byte("b"))
	_ = e.SaveEvent("e3", []byte("c"))

	all, err := e.ListEvents()
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("ListEvents: got %d entries, want 3", len(all))
	}

	for _, id := range []string{"e1", "e2", "e3"} {
		if _, ok := all[id]; !ok {
			t.Errorf("ListEvents: missing id %q", id)
		}
	}
}

func TestEventStore_ListEvents_ExcludesStatusKeys(t *testing.T) {
	e := newTestEventStore()

	_ = e.SaveEvent("e1", []byte("data"))
	_ = e.MarkProcessed("e1")

	all, err := e.ListEvents()
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	// Must have exactly one entry (the payload), not two (payload + status).
	if len(all) != 1 {
		t.Errorf("ListEvents: got %d entries, want 1 (status key must be hidden)", len(all))
	}
}

func TestEventStore_ListEvents_Empty(t *testing.T) {
	e := newTestEventStore()

	all, err := e.ListEvents()
	if err != nil {
		t.Fatalf("ListEvents empty: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("ListEvents empty: got %d entries, want 0", len(all))
	}
}

func TestEventStore_ListEvents_IsolatedFromSessions(t *testing.T) {
	backend := gocache.NewBackend(time.Minute)
	sessions := gocache.NewSessionStore(backend)
	events := gocache.NewEventStore(backend)

	_ = sessions.SaveSession("s1", []byte("session"))
	_ = events.SaveEvent("e1", []byte("event"))

	all, err := events.ListEvents()
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("ListEvents should not include session keys; got %d entries", len(all))
	}
}
