package cache

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func newCache(t *testing.T) *Cache {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := Open(path, time.Hour)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestPutGetRoundTrip(t *testing.T) {
	c := newCache(t)
	want := sample{Name: "rom-8", Count: 39}
	if err := c.Put("k", want); err != nil {
		t.Fatalf("Put: %v", err)
	}
	var got sample
	found, err := c.Get("k", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected hit")
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestGetMiss(t *testing.T) {
	c := newCache(t)
	var got sample
	found, err := c.Get("missing", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected miss")
	}
}

func TestExpiredEntrySelfDeletes(t *testing.T) {
	c := newCache(t)
	if err := c.PutWithTTL("k", sample{Name: "x"}, -time.Second); err != nil {
		t.Fatalf("Put: %v", err)
	}
	var got sample
	found, err := c.Get("k", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected expired miss")
	}
	// Confirm the entry was removed from disk.
	err = c.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketCache).Get([]byte("k")); v != nil {
			return errors.New("entry still present")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected entry deleted: %v", err)
	}
}

func TestPoisonedEntrySelfDeletes(t *testing.T) {
	c := newCache(t)
	if err := c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCache).Put([]byte("k"), []byte("not json"))
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var got sample
	found, err := c.Get("k", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected miss on poisoned entry")
	}
	err = c.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketCache).Get([]byte("k")); v != nil {
			return errors.New("poisoned entry still present")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected entry deleted: %v", err)
	}
}

func TestSchemaMismatchClearsCacheBucket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := Open(path, time.Hour)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := c.Put("k", sample{Name: "preserve me", Count: 1}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Tamper with the on-disk schema version.
	tampered, err := bolt.Open(path, 0644, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("reopen for tamper: %v", err)
	}
	if err := tampered.Update(func(tx *bolt.Tx) error {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], 99)
		return tx.Bucket(bucketMeta).Put(keySchema, buf[:])
	}); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := tampered.Close(); err != nil {
		t.Fatalf("close tampered: %v", err)
	}

	c2, err := Open(path, time.Hour)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	defer c2.Close()
	var got sample
	found, err := c2.Get("k", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected cache bucket to be cleared on schema mismatch")
	}
	// DB file is preserved (still openable, meta bucket intact).
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("expected DB preserved: %v", statErr)
	}
}

func TestCorruptDBRenamedAside(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.db")
	if err := os.WriteFile(path, []byte("not a bolt db at all"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	c, err := Open(path, time.Hour)
	if err != nil {
		t.Fatalf("Open should recover: %v", err)
	}
	defer c.Close()
	// Confirm a .broken-* sibling exists.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	foundBroken := false
	for _, e := range entries {
		name := e.Name()
		if len(name) > len("cache.db.broken-") && name[:len("cache.db.broken-")] == "cache.db.broken-" {
			foundBroken = true
			break
		}
	}
	if !foundBroken {
		t.Fatal("expected a cache.db.broken-* file to exist")
	}
}

func TestClearEmptiesBucketPreservesSchema(t *testing.T) {
	c := newCache(t)
	if err := c.Put("k", sample{Name: "x"}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	var got sample
	found, _ := c.Get("k", &got)
	if found {
		t.Fatal("expected miss after Clear")
	}
	// Schema bytes still present.
	err := c.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketMeta).Get(keySchema); v == nil {
			return errors.New("schema gone")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("schema check: %v", err)
	}
}

func TestSweepCountsAndDeletes(t *testing.T) {
	c := newCache(t)
	if err := c.PutWithTTL("expired1", sample{Name: "a"}, -time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.PutWithTTL("expired2", sample{Name: "b"}, -time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.Put("fresh", sample{Name: "c"}); err != nil {
		t.Fatal(err)
	}
	n, err := c.Sweep()
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if n != 2 {
		t.Fatalf("Sweep count = %d, want 2", n)
	}
	var got sample
	if found, _ := c.Get("fresh", &got); !found {
		t.Fatal("expected fresh entry to survive")
	}
}

func TestStats(t *testing.T) {
	c := newCache(t)
	if err := c.Put("a", sample{Name: "1"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Put("b", sample{Name: "2"}); err != nil {
		t.Fatal(err)
	}
	s, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if s.Entries != 2 {
		t.Fatalf("entries = %d, want 2", s.Entries)
	}
	if s.SizeBytes <= 0 {
		t.Fatalf("size = %d, want > 0", s.SizeBytes)
	}
	if s.Oldest.IsZero() {
		t.Fatal("oldest should be set")
	}
}
