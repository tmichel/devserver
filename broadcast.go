package main

import (
	"sync"
)

// Broadcaster is a generic type that allows broadcasting messages of type T
// to multiple listeners.
type Broadcaster[T any] struct {
	listeners map[*chan T]struct{}
	mu        sync.RWMutex
}

// NewBroadcaster creates and returns a new Broadcaster for type T.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		listeners: make(map[*chan T]struct{}),
	}
}

// AddListener creates a new channel for receiving broadcast messages and
// registers it with the Broadcaster. It returns the newly created receive-only channel
// and a function to remove the listener.
func (b *Broadcaster[T]) AddListener() (<-chan T, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan T)
	b.listeners[&ch] = struct{}{}

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.listeners, &ch)
		close(ch)
	}
}

// Broadcast sends the given message to all registered listeners.
// This operation is non-blocking for the broadcaster, but may block
// if any listener's channel is full.
func (b *Broadcaster[T]) Broadcast(message T) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.listeners {
		go func(c *chan T) {
			*c <- message
		}(ch)
	}
}
