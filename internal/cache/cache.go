// Package cache provides a TTL-bounded, JSON-encoded key/value store backed
// by bbolt. It is used by the lookup decorator to memoize scripture and
// translation API responses on disk.
//
// SchemaVersion
//
// The on-disk format is versioned via SchemaVersion. Bump it whenever any
// of the following change in a way that existing cached payloads cannot be
// safely decoded by the new code:
//
//   - field type change, rename, or removal in cached payload types
//     (e.g. fields on lookup.LookupResult, lookup.Verse, lookup.Translation,
//     or anything the decorator stores under its own envelope);
//   - the key format (e.g. moving from "v1|..." to "v2|...");
//   - the Entry envelope shape.
//
// Adding new optional fields does NOT require a bump — JSON tolerates
// missing fields on decode.
//
// On schema mismatch the cache bucket is dropped and recreated; the
// underlying DB file is preserved. Per-entry decode failures self-delete
// the bad entry. Corrupt DB files on open are renamed aside
// (<path>.broken-<unix>) and a fresh DB is opened in their place. The cache
// never panics and never blocks the application.
package cache

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

// SchemaVersion is the on-disk format version. See package doc for bump rules.
const SchemaVersion uint32 = 1

var (
	bucketMeta  = []byte("meta")
	bucketCache = []byte("cache")
	keySchema   = []byte("schema_version")
)

// Entry is the on-disk envelope around a cached value. Payload is the
// JSON-encoded user value; the envelope itself is also JSON.
type Entry struct {
	FetchedAt time.Time       `json:"fetched_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	Payload   json.RawMessage `json:"payload"`
}

// Cache is a bbolt-backed KV store for lookup results.
type Cache struct {
	db  *bolt.DB
	ttl time.Duration
}

// Stats describes the current state of the cache.
type Stats struct {
	Entries   int       `json:"entries"`
	SizeBytes int64     `json:"size_bytes"`
	Oldest    time.Time `json:"oldest"`
}

// Open opens (or creates) the cache file at the given path with the given
// default TTL. Corrupt files are renamed aside and a fresh DB is opened.
// Schema mismatches drop and recreate the cache bucket.
func Open(path string, ttl time.Duration) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("cache: mkdir: %w", err)
	}
	db, err := openOrRecover(path)
	if err != nil {
		return nil, err
	}
	c := &Cache{db: db, ttl: ttl}
	if err := c.initBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return c, nil
}

// openOrRecover opens the bolt DB. If the existing file is corrupt or
// otherwise unopenable, it is renamed to <path>.broken-<unix> and a fresh
// DB is opened in its place.
func openOrRecover(path string) (*bolt.DB, error) {
	opts := &bolt.Options{Timeout: 2 * time.Second}
	db, err := bolt.Open(path, 0644, opts)
	if err == nil {
		return db, nil
	}
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("cache: open: %w", err)
	}
	broken := fmt.Sprintf("%s.broken-%d", path, time.Now().Unix())
	if renameErr := os.Rename(path, broken); renameErr != nil {
		return nil, fmt.Errorf("cache: open failed (%v) and rename aside failed: %w", err, renameErr)
	}
	db, err2 := bolt.Open(path, 0644, opts)
	if err2 != nil {
		return nil, fmt.Errorf("cache: open after recovery failed: %w", err2)
	}
	return db, nil
}

// initBuckets ensures both buckets exist and reconciles the schema version.
// If the on-disk schema differs from SchemaVersion the cache bucket is
// dropped and recreated.
func (c *Cache) initBuckets() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		meta, err := tx.CreateBucketIfNotExists(bucketMeta)
		if err != nil {
			return err
		}
		stored := uint32(0)
		if v := meta.Get(keySchema); len(v) == 4 {
			stored = binary.BigEndian.Uint32(v)
		}
		if stored != SchemaVersion {
			if stored != 0 {
				fmt.Printf("cache: schema %d \u2192 %d; clearing cache bucket\n", stored, SchemaVersion)
			}
			if err := tx.DeleteBucket(bucketCache); err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
				return err
			}
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], SchemaVersion)
			if err := meta.Put(keySchema, buf[:]); err != nil {
				return err
			}
		}
		_, err = tx.CreateBucketIfNotExists(bucketCache)
		return err
	})
}

// Close closes the underlying DB.
func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}

// SetTTL updates the default TTL used by Put. Existing entries are unaffected.
func (c *Cache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Get fetches a cached value and decodes its payload into out. Returns
// (false, nil) for misses, expired entries, or undecodable payloads.
// Expired or undecodable entries are deleted as a side effect.
func (c *Cache) Get(key string, out any) (bool, error) {
	if c == nil || c.db == nil {
		return false, nil
	}
	var entry Entry
	var found bool
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return nil
		}
		raw := b.Get([]byte(key))
		if raw == nil {
			return nil
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			found = false
			return errPoisoned
		}
		found = true
		return nil
	})
	if errors.Is(err, errPoisoned) {
		_ = c.Delete(key)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		_ = c.Delete(key)
		return false, nil
	}
	if err := json.Unmarshal(entry.Payload, out); err != nil {
		_ = c.Delete(key)
		return false, nil
	}
	return true, nil
}

var errPoisoned = errors.New("cache: poisoned entry")

// Put stores value under key using the cache's default TTL.
func (c *Cache) Put(key string, value any) error {
	return c.PutWithTTL(key, value, c.ttl)
}

// PutWithTTL stores value under key with the given TTL. A zero TTL stores
// the entry with no expiration. Negative TTLs produce entries that have
// already expired (useful in tests).
func (c *Cache) PutWithTTL(key string, value any, ttl time.Duration) error {
	if c == nil || c.db == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal payload: %w", err)
	}
	now := time.Now()
	entry := Entry{
		FetchedAt: now,
		Payload:   payload,
	}
	if ttl != 0 {
		entry.ExpiresAt = now.Add(ttl)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("cache: marshal entry: %w", err)
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return errors.New("cache: bucket missing")
		}
		return b.Put([]byte(key), data)
	})
}

// Delete removes a single entry. Missing keys are not an error.
func (c *Cache) Delete(key string) error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

// Clear drops and recreates the cache bucket, leaving meta intact.
func (c *Cache) Clear() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketCache); err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return err
		}
		_, err := tx.CreateBucket(bucketCache)
		return err
	})
}

// Sweep deletes every expired entry and returns how many were removed.
func (c *Cache) Sweep() (int, error) {
	if c == nil || c.db == nil {
		return 0, nil
	}
	now := time.Now()
	var toDelete [][]byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var entry Entry
			if err := json.Unmarshal(v, &entry); err != nil {
				key := make([]byte, len(k))
				copy(key, k)
				toDelete = append(toDelete, key)
				return nil
			}
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				key := make([]byte, len(k))
				copy(key, k)
				toDelete = append(toDelete, key)
			}
			return nil
		})
	})
	if err != nil {
		return 0, err
	}
	if len(toDelete) == 0 {
		return 0, nil
	}
	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return nil
		}
		for _, k := range toDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(toDelete), nil
}

// Stats returns aggregate counters for the cache.
func (c *Cache) Stats() (Stats, error) {
	var s Stats
	if c == nil || c.db == nil {
		return s, nil
	}
	if info, err := os.Stat(c.db.Path()); err == nil {
		s.SizeBytes = info.Size()
	}
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCache)
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, v []byte) error {
			s.Entries++
			var entry Entry
			if err := json.Unmarshal(v, &entry); err != nil {
				return nil
			}
			if s.Oldest.IsZero() || entry.FetchedAt.Before(s.Oldest) {
				s.Oldest = entry.FetchedAt
			}
			return nil
		})
	})
	return s, err
}
