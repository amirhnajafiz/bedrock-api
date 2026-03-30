package storage

import "errors"

// ErrNotFound is returned when a requested key does not exist in the store.
var ErrNotFound = errors.New("storage: key not found")