package domain

import (
	"context"
	"sync"
	"time"
)

// OTPCache interface for OTP caching
type OTPCache interface {
	Put(ctx context.Context, email, otp string, ttl time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

// InMemoryOTPCache is a simple in-memory implementation of OTPCache
type InMemoryOTPCache struct {
	mu    sync.RWMutex
	store map[string]*otpEntry
}

type otpEntry struct {
	otp       string
	expiresAt time.Time
}

// NewInMemoryOTPCache creates a new in-memory OTP cache
func NewInMemoryOTPCache() *InMemoryOTPCache {
	cache := &InMemoryOTPCache{
		store: make(map[string]*otpEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Put stores OTP in cache with TTL
func (c *InMemoryOTPCache) Put(ctx context.Context, email, otp string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[email] = &otpEntry{
		otp:       otp,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Get retrieves OTP from cache
func (c *InMemoryOTPCache) Get(ctx context.Context, email string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.store[email]
	if !exists {
		return "", nil
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return "", nil
	}

	return entry.otp, nil
}

// Delete removes OTP from cache
func (c *InMemoryOTPCache) Delete(ctx context.Context, email string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, email)
	return nil
}

// cleanup removes expired entries periodically
func (c *InMemoryOTPCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for email, entry := range c.store {
			if now.After(entry.expiresAt) {
				delete(c.store, email)
			}
		}
		c.mu.Unlock()
	}
}
