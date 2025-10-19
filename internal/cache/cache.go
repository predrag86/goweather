package cache

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"time"

	"goweather/internal/log"
)

type Item struct {
	Data      any
	Timestamp time.Time
}

type Cache struct {
	mu        sync.RWMutex
	items     map[string]Item
	expiry    time.Duration
	cacheFile string
}

// NewCache creates a cache with given expiration time (e.g. 10min).
func NewCache(expiry time.Duration) *Cache {
	dir, _ := os.UserCacheDir()
	path := filepath.Join(dir, "goweather", "weather_cache.gob")
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	c := &Cache{
		items:     make(map[string]Item),
		expiry:    expiry,
		cacheFile: path,
	}
	c.loadFromFile()
	log.Logger.Infow("Cache initialized",
		"path", c.cacheFile,
		"expiry", expiry.String(),
	)
	return c
}

// Get retrieves a valid cached item (if not expired).
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		log.Logger.Debugw("Cache miss", "key", key)
		return nil, false
	}

	if time.Since(item.Timestamp) > c.expiry {
		log.Logger.Infow("Cache expired", "key", key,
			"age", time.Since(item.Timestamp).Round(time.Second).String())
		delete(c.items, key)
		return nil, false
	}

	log.Logger.Debugw("Cache hit", "key", key)
	return item.Data, true
}

// Set stores a new item and persists to disk.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item{Data: value, Timestamp: time.Now()}
	c.saveToFile()
	log.Logger.Infow("Cache updated",
		"key", key,
		"total_items", len(c.items),
	)
}

// BackgroundRefresh launches a goroutine that refreshes a key periodically.
func (c *Cache) BackgroundRefresh(key string, refreshFn func() (any, error)) {
	go func() {
		for {
			time.Sleep(c.expiry)
			data, err := refreshFn()
			if err != nil {
				log.Logger.Warnw("Background refresh failed", "key", key, "error", err)
				continue
			}
			c.Set(key, data)
			log.Logger.Infow("Cache refreshed in background", "key", key)
		}
	}()
}

// --- internal persistence helpers ---

func (c *Cache) loadFromFile() {
	file, err := os.Open(c.cacheFile)
	if err != nil {
		log.Logger.Debugw("No existing cache file, starting empty", "path", c.cacheFile)
		return
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	if err := dec.Decode(&c.items); err != nil {
		log.Logger.Warnw("Failed to decode cache", "error", err)
	} else {
		log.Logger.Infow("Cache loaded from disk",
			"items", len(c.items),
			"path", c.cacheFile,
		)
	}
}

func (c *Cache) saveToFile() {
	file, err := os.Create(c.cacheFile)
	if err != nil {
		log.Logger.Warnw("Failed to save cache", "error", err)
		return
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(c.items); err != nil {
		log.Logger.Warnw("Failed to encode cache", "error", err)
	} else {
		log.Logger.Debugw("Cache saved to disk", "path", c.cacheFile)
	}
}
