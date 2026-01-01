package op

import (
	"errors"
	"sync"
	"time"
)

// Rate limit errors - distinct for RPM vs RPD
var (
	ErrRateLimitRPM = errors.New("rate limit exceeded: too many requests per minute")
	ErrRateLimitRPD = errors.New("rate limit exceeded: too many requests per day")
)

// rateLimitEntry tracks request counts for an API key
type rateLimitEntry struct {
	minuteCount   int       // requests in current minute window
	minuteResetAt time.Time // when minute window resets
	dayCount      int       // requests in current day window
	dayResetAt    time.Time // when day window resets
}

var (
	rateLimitCache     = make(map[int]*rateLimitEntry)
	rateLimitCacheLock sync.RWMutex
)

// RateLimitCheck checks if the API key has exceeded rate limits.
// Returns nil if allowed, ErrRateLimitRPM or ErrRateLimitRPD if exceeded.
// If rpm or rpd is 0, that limit is unlimited.
func RateLimitCheck(apiKeyID int, rpm int, rpd int) error {
	// If both unlimited, skip check
	if rpm == 0 && rpd == 0 {
		return nil
	}

	now := time.Now()

	rateLimitCacheLock.Lock()
	defer rateLimitCacheLock.Unlock()

	entry, exists := rateLimitCache[apiKeyID]
	if !exists {
		entry = &rateLimitEntry{
			minuteResetAt: now.Add(time.Minute),
			dayResetAt:    getNextDayReset(now),
		}
		rateLimitCache[apiKeyID] = entry
	}

	// Reset minute window if expired
	if now.After(entry.minuteResetAt) {
		entry.minuteCount = 0
		entry.minuteResetAt = now.Add(time.Minute)
	}

	// Reset day window if expired
	if now.After(entry.dayResetAt) {
		entry.dayCount = 0
		entry.dayResetAt = getNextDayReset(now)
	}

	// Check RPM limit (if set)
	if rpm > 0 && entry.minuteCount >= rpm {
		return ErrRateLimitRPM
	}

	// Check RPD limit (if set)
	if rpd > 0 && entry.dayCount >= rpd {
		return ErrRateLimitRPD
	}

	return nil
}

// RateLimitIncrement increments the request counters for an API key.
// Should be called after a successful request.
func RateLimitIncrement(apiKeyID int) {
	now := time.Now()

	rateLimitCacheLock.Lock()
	defer rateLimitCacheLock.Unlock()

	entry, exists := rateLimitCache[apiKeyID]
	if !exists {
		entry = &rateLimitEntry{
			minuteResetAt: now.Add(time.Minute),
			dayResetAt:    getNextDayReset(now),
		}
		rateLimitCache[apiKeyID] = entry
	}

	// Reset windows if needed before incrementing
	if now.After(entry.minuteResetAt) {
		entry.minuteCount = 0
		entry.minuteResetAt = now.Add(time.Minute)
	}
	if now.After(entry.dayResetAt) {
		entry.dayCount = 0
		entry.dayResetAt = getNextDayReset(now)
	}

	entry.minuteCount++
	entry.dayCount++
}

// getNextDayReset returns the start of the next day (midnight)
func getNextDayReset(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

// RateLimitCleanup removes entries for deleted API keys
func RateLimitCleanup(apiKeyID int) {
	rateLimitCacheLock.Lock()
	defer rateLimitCacheLock.Unlock()
	delete(rateLimitCache, apiKeyID)
}
