package framework

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/gofrs/flock"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

const (
	lockFilePath = "moley.lock"
)

// LockEntry represents a persisted resource snapshot in moley.lock.
type LockEntry struct {
	Key         string `json:"key"`
	Data        any    `json:"data"`
	HandlerName string `json:"handler_name"`
	InputHash   string `json:"input_hash,omitempty"`
}

// LockFile manages persistent storage of resource snapshots in moley.lock.
type LockFile struct {
	Entries []LockEntry `json:"entries"`
	flock   *flock.Flock
}

// LoadLockFile loads moley.lock from disk and acquires an exclusive file lock.
// Returns an empty LockFile if the file is missing or corrupt.
func LoadLockFile() (*LockFile, error) {
	fl := flock.New(lockFilePath)

	if err := fl.Lock(); err != nil {
		return nil, fmt.Errorf("failed to acquire file lock: %w", err)
	}

	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{flock: fl}, nil
		}
		_ = fl.Unlock()
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	lf := &LockFile{flock: fl}
	if err := json.Unmarshal(data, lf); err != nil {
		logger.Warnf("Lock file is corrupt, starting fresh (resources will be rediscovered)", map[string]any{
			"error": err.Error(),
		})
		return &LockFile{flock: fl}, nil
	}

	return lf, nil
}

// Close releases the file lock.
func (lf *LockFile) Close() error {
	if lf.flock == nil {
		return nil
	}
	err := lf.flock.Unlock()
	lf.flock = nil
	return err
}

// Save persists the lock file to disk.
// Safe to write directly — the exclusive flock prevents concurrent access.
func (lf *LockFile) Save() error {
	data, err := json.Marshal(lf)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(lockFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// PurgeOrphans removes lock entries whose handler name is not in the registered set.
func (lf *LockFile) PurgeOrphans(registeredHandlers map[string]bool) error {
	before := len(lf.Entries)
	lf.Entries = slices.DeleteFunc(lf.Entries, func(e LockEntry) bool {
		return !registeredHandlers[e.HandlerName]
	})
	after := len(lf.Entries)

	if before != after {
		logger.Infof("Purged orphaned lock entries", map[string]any{
			"purged": before - after,
		})
		return lf.Save()
	}
	return nil
}
