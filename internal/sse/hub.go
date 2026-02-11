package sse

import (
	"encoding/json"
	"sync"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint64]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint64]map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe(userID uint64) (chan Event, func()) {
	ch := make(chan Event, 16)

	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[chan Event]struct{})
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()

	cleanup := func() {
		h.mu.Lock()
		delete(h.subscribers[userID], ch)
		if len(h.subscribers[userID]) == 0 {
			delete(h.subscribers, userID)
		}
		h.mu.Unlock()
		close(ch)
	}

	return ch, cleanup
}

func (h *Hub) Publish(userID uint64, event Event) {
	h.mu.RLock()
	channels := h.subscribers[userID]
	h.mu.RUnlock()

	for ch := range channels {
		select {
		case ch <- event:
		default:
			// Drop if buffer full
		}
	}
}

func (h *Hub) PublishToMany(userIDs []uint64, event Event) {
	for _, uid := range userIDs {
		h.Publish(uid, event)
	}
}

func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
