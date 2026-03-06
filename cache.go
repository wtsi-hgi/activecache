/*******************************************************************************
 * Copyright (c) 2026 Genome Research Ltd.
 *
 * Author: Michael Woolnough <mw31@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

// Package activecache implements a generic, self-populating cache, that
// automatically refreshes stored data.
package activecache

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"
)

// Cache wraps a function, caching requested information, and re-retrieving on a
// schedule.
type Cache[K comparable, V any] struct {
	stop  func()
	getFn func(K) (V, error)

	mu    sync.RWMutex
	cache map[K]value[V]
}

type value[V any] struct {
	v   V
	err error
}

// New creates a cache storing the last response from a function, given a key.
//
// The information will be re-retrieve on a timeout specified by the given
// Duration.
//
// If a re-retrieval produces an error, it will not overwrite the existing
// value.
//
// The Stop() method must be before replacing (or otherwise losing this pointer
// to) this cache.
func New[K comparable, V any](d time.Duration, getFn func(K) (V, error)) *Cache[K, V] {
	ctx, fn := context.WithCancel(context.Background())

	cache := &Cache[K, V]{
		cache: make(map[K]value[V]),
		stop:  fn,
		getFn: getFn,
	}

	if d > 0 {
		go cache.runCache(ctx, d)
	}

	return cache
}

func (c *Cache[K, V]) runCache(ctx context.Context, d time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
		}

		c.mu.RLock()
		keys := slices.Collect(maps.Keys(c.cache))
		c.mu.RUnlock()

		updates := make(map[K]value[V])

		for _, key := range keys {
			v, err := c.getFn(key)
			if err != nil {
				if oldV, oldErr := c.Get(key); oldErr == nil {
					v = oldV
					err = nil
				}
			}

			updates[key] = value[V]{v, err}
		}

		c.mu.Lock()
		maps.Copy(c.cache, updates)
		c.mu.Unlock()
	}
}

// Get attempts to retrieve a cached item using the given key.
//
// If the item doesn't exist, it will call the function given to 'New' to get an
// item, then cache it and return it.
func (c *Cache[K, V]) Get(key K) (V, error) {
	c.mu.RLock()
	existing, ok := c.cache[key]
	c.mu.RUnlock()

	if ok {
		return existing.v, existing.err
	}

	t, err := c.getFn(key)

	c.mu.Lock()
	c.cache[key] = value[V]{t, err}
	c.mu.Unlock()

	return t, err
}

// Remove removes the key from the cache, returning true if the key existed in
// the cache.
func (c *Cache[K, V]) Remove(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.cache[key]

	delete(c.cache, key)

	return ok
}

// Stop stops the re-generating of cached items.
func (c *Cache[K, V]) Stop() {
	c.stop()
}
