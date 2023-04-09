package evecache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cockroachdb/pebble"
)

// New returns a new Cache that will store items in an in-memory map and on disk in pebbledb.
func New(cachDir string) (*EveCache, error) {
	pebDbPath, err := db_location(cachDir)
	if err != nil {
		return nil, err
	}
	fmt.Println("Storing http cache on disk in pebbledb at", pebDbPath)
	pdb, err := pebble.Open(pebDbPath, &pebble.Options{})
	if err != nil {
		return nil, err
	}

	c := &EveCache{
		pebbledb: pdb,
		items:    map[string][]byte{},
	}
	return c, nil
}

// EveCache is an implementation of Cache that stores responses in an in-memory map.
type EveCache struct {
	pebbledb *pebble.DB

	mu    sync.RWMutex
	items map[string][]byte
}

// Get returns the []byte representation of the response and true if present, false if not
func (c *EveCache) Get(key string) (resp []byte, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, ok = c.items[key]

	if ok {
		return resp, ok
	}
	value, closer, err := c.pebbledb.Get([]byte(key))
	if err == pebble.ErrNotFound {
		return nil, false
	}
	if err != nil {
		panic(fmt.Sprintf("Error while trying to get key %s from pebble: %s", key, err))
	}
	dst := make([]byte, len(value))
	copy(dst, value)
	if err := closer.Close(); err != nil {
		panic(fmt.Sprintf("Error while trying to close pebble: %s", err))
	}

	return dst, true
}

// Set saves response resp to the cache with key
func (c *EveCache) Set(key string, resp []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = resp

	// Because this is a cache, we don't need to sync (pebble.Sync) the data to disk.
	// Also calling *EveCache.Close() will sync the data to disk.
	err := c.pebbledb.Set([]byte(key), resp, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Error while trying to set key %s to pebble: %s", key, err))
	}
}

// Delete removes key from the cache
func (c *EveCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)

	err := c.pebbledb.Delete([]byte(key), pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Error while trying to delete key %s from pebble: %s", key, err))
	}
}

// Close closes the cache and flushes any pending writes to disk
func (c *EveCache) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.pebbledb.Close()
}

func db_location(baseDir string) (string, error) {
	// tmpDir := os.TempDir()
	// myTmpDir := filepath.Join(tmpDir, "evecache_tmp")
	dbpath := filepath.Join(baseDir, "evecache_peb_db")

	_, err := os.Stat(dbpath)
	if os.IsNotExist(err) {
		err := os.Mkdir(dbpath, 0700)
		if err != nil {
			return "", fmt.Errorf("could not create directory %s: %w", dbpath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat directory %s: %w", dbpath, err)
	}

	return dbpath, nil
}
