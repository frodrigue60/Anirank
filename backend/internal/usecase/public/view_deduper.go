package public

import (
	"sync"
	"time"
)

type viewDeduper struct {
	mu          sync.Mutex
	ttl         time.Duration
	entries     map[string]time.Time
	nextCleanup time.Time
}

func newViewDeduper(ttl time.Duration) *viewDeduper {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &viewDeduper{
		ttl:     ttl,
		entries: make(map[string]time.Time),
	}
}

// MarkIfNew atomically reserves a view key until its TTL expires.
// It returns false when the key is still inside the deduplication window.
func (d *viewDeduper) MarkIfNew(key string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if expiresAt, ok := d.entries[key]; ok && now.Before(expiresAt) {
		return false
	}

	if d.nextCleanup.IsZero() || !now.Before(d.nextCleanup) {
		for entryKey, expiresAt := range d.entries {
			if !now.Before(expiresAt) {
				delete(d.entries, entryKey)
			}
		}
		d.nextCleanup = now.Add(time.Hour)
	}

	d.entries[key] = now.Add(d.ttl)
	return true
}

func (d *viewDeduper) SetExpiry(key string, expiresAt, now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !expiresAt.After(now) {
		delete(d.entries, key)
		return
	}
	d.entries[key] = expiresAt
}
