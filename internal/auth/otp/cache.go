package otp

import (
	"sync"
	"time"
)

// Cache defines temporary OTP storage behavior.
type Cache interface {
	Set(email, otp string, ttl time.Duration)
	Get(email string) (string, bool)
	Delete(email string)
}

type entry struct {
	value     string
	expiresAt time.Time
}

// InMemoryCache is a goroutine-safe in-memory implementation.
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]entry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{items: make(map[string]entry)}
}

func (c *InMemoryCache) Set(email, otp string, ttl time.Duration) {
	c.mu.Lock()
	c.items[email] = entry{value: otp, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *InMemoryCache) Get(email string) (string, bool) {
	c.mu.RLock()
	item, ok := c.items[email]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(item.expiresAt) {
		c.Delete(email)
		return "", false
	}
	return item.value, true
}

func (c *InMemoryCache) Delete(email string) {
	c.mu.Lock()
	delete(c.items, email)
	c.mu.Unlock()
}
