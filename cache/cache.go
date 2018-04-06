package cache

import (
	"sync"
	"time"
	// r "gopkg.in/redis.v5"
)

//Storage mecanism for caching strings
type Storage interface {
	Get(key string) []byte
	Set(key string, content []byte, duration time.Duration)
}

// const preffix = "_PAGE_CACHE_"
//
// // RedisCache storage mecanism for caching strings in memory
// type RedisCache struct {
// 	client *r.Client
// }
//
// // NewRedisCache creates a new redis storage
// func NewRedisCache(url string) (*RedisCache, error) {
// 	var (
// 		opts *r.Options
// 		err  error
// 	)
//
// 	if opts, err = r.ParseURL(url); err != nil {
// 		return nil, err
// 	}
//
// 	return &RedisCache{
// 		client: r.NewClient(opts),
// 	}, nil
// }
//
// // Get a cached content by key
// func (c RedisCache) Get(key string) []byte {
// 	val, _ := c.client.Get(preffix + key).Bytes()
// 	return val
// }
//
// // Set a cached content by key
// func (c RedisCache) Set(key string, content []byte, duration time.Duration) {
// 	c.client.Set(preffix+key, content, duration)
// }

// MEMORY CACHE

// Item is a cached reference
type Item struct {
	Content    []byte
	Expiration int64
}

// Expired returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// MemoryCache mecanism for caching strings in memory
type MemoryCache struct {
	items map[string]Item
	mu    *sync.RWMutex
}

// NewMemoryCache creates a new in memory storage
func NewMemoryCache() Storage {
	return &MemoryCache{
		items: make(map[string]Item),
		mu:    &sync.RWMutex{},
	}
}

// Get a cached content by key
func (c MemoryCache) Get(key string) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item := c.items[key]
	if item.Expired() {
		delete(c.items, key)
		return nil
	}
	return item.Content
}

// Set a cached content by key
func (c MemoryCache) Set(key string, content []byte, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Content:    content,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}
