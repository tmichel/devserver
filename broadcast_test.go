package main

import (
	"sync"
	"testing"
	"time"
)

func TestBroadcaster(t *testing.T) {
	t.Run("AddListener", func(t *testing.T) {
		b := NewBroadcaster[int]()
		ch, remove := b.AddListener()
		if len(b.listeners) != 1 {
			t.Errorf("Expected 1 listener, got %d", len(b.listeners))
		}
		remove()
		if len(b.listeners) != 0 {
			t.Errorf("Expected 0 listeners after removal, got %d", len(b.listeners))
		}
		_, ok := <-ch
		if ok {
			t.Error("Channel should be closed after removal")
		}
	})

	t.Run("Broadcast", func(t *testing.T) {
		b := NewBroadcaster[int]()
		ch1, _ := b.AddListener()
		ch2, _ := b.AddListener()

		b.Broadcast(42)

		for i, ch := range []<-chan int{ch1, ch2} {
			select {
			case msg := <-ch:
				if msg != 42 {
					t.Errorf("Listener %d: Expected 42, got %d", i, msg)
				}
			case <-time.After(time.Second):
				t.Errorf("Timeout waiting for message on listener %d", i)
			}
		}
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		b := NewBroadcaster[int]()
		var wg sync.WaitGroup
		listenerCount := 100
		messageCount := 100

		// Add listeners concurrently
		for i := 0; i < listenerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				b.AddListener()
			}()
		}
		wg.Wait()

		if len(b.listeners) != listenerCount {
			t.Errorf("Expected %d listeners, got %d", listenerCount, len(b.listeners))
		}

		// Prepare a channel to signal completion of broadcasting
		done := make(chan struct{})

		// Broadcast messages concurrently
		go func() {
			var broadcastWg sync.WaitGroup
			for i := 0; i < messageCount; i++ {
				broadcastWg.Add(1)
				go func(msg int) {
					defer broadcastWg.Done()
					b.Broadcast(msg)
				}(i)
			}
			broadcastWg.Wait()
			close(done)
		}()

		// Wait for broadcasting to complete or timeout
		select {
		case <-done:
			// Broadcasting completed successfully
		case <-time.After(time.Second):
			t.Fatal("Broadcasting timed out after 1 second")
		}
	})

	t.Run("RemoveListener", func(t *testing.T) {
		b := NewBroadcaster[int]()
		_, remove1 := b.AddListener()
		ch2, _ := b.AddListener()

		remove1()
		b.Broadcast(42)

		select {
		case msg := <-ch2:
			if msg != 42 {
				t.Errorf("Expected 42, got %d on ch2", msg)
			}
		case <-time.After(time.Second):
			t.Error("Timeout waiting for message on ch2")
		}
	})

	t.Run("AllListenersReceiveMessage", func(t *testing.T) {
		b := NewBroadcaster[int]()
		listenerCount := 10
		listeners := make([]<-chan int, listenerCount)

		for i := 0; i < listenerCount; i++ {
			listeners[i], _ = b.AddListener()
		}

		b.Broadcast(42)

		for i, ch := range listeners {
			select {
			case msg := <-ch:
				if msg != 42 {
					t.Errorf("Listener %d: Expected 42, got %d", i, msg)
				}
			case <-time.After(time.Second):
				t.Errorf("Timeout waiting for message on listener %d", i)
			}
		}
	})
}
