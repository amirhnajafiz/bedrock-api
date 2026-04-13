package storage_test

import (
	"testing"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
)

// TestGoCacheStorage tests the basic functionality of the go-cache storage backend.
func TestGoCacheStorageSet(t *testing.T) {
	// create a new instance of the go-cache storage backend
	storage := storage.NewGoCache()

	// test setting and getting a value
	key := "testKey"
	value := []byte("testValue")

	if err := storage.Set(key, value); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrievedValue, err := storage.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(retrievedValue) != string(value) {
		t.Fatalf("Expected value %s, got %s", value, retrievedValue)
	}

	// test deleting a value
	if err := storage.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Get(key)
	if err == nil {
		t.Fatalf("Expected error when getting deleted key, got nil")
	}

	// test listing values with a wildcard
	if err := storage.Set("prefix1", []byte("value1")); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := storage.Set("prefix2", []byte("value2")); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	values, err := storage.List("prefix")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
}
