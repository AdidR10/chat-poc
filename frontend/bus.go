package main

import "sync"

var appBus *Bus

// // Event struct can hold any bus event
// type Event struct {
// 	Type string
// 	Data interface{}
// }

type Event struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
}

type Bus struct {
	subscribers []chan Event
	lock        sync.Mutex
}

func NewBus() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe() <-chan Event {
	b.lock.Lock()
	defer b.lock.Unlock()
	ch := make(chan Event, 20) // buffered channel
	b.subscribers = append(b.subscribers, ch)
	return ch
}

func (b *Bus) Publish(event Event) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default: // avoid blocking
		}
	}
}