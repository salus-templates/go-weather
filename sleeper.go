package main

import (
	"log"
	"time"
)

// Sleeper interface defines the contract for sleeping.
type Sleeper interface {
	Sleep(d time.Duration)
}

// DefaultSleeper implements Sleeper using time.Sleep.
type DefaultSleeper struct{}

// Sleep pauses the current goroutine for at least the duration d.
func (s *DefaultSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}

// NoOpSleeper implements Sleeper but does nothing.
// This is useful for tests where you don't want actual delays.
type NoOpSleeper struct{}

// Sleep does nothing.
func (s *NoOpSleeper) Sleep(d time.Duration) {
	log.Printf("NoOpSleeper: sleep called for %v\n", d)
	// No operation, effectively zero delay
}
