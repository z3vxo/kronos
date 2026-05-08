package broker

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func NewBroker() *Broker {
	return &Broker{
		Channels: make(map[string]chan string),
	}
}

type Broker struct {
	Channels map[string]chan string
	mu       sync.RWMutex
}

func (b *Broker) AddSubscriber() (string, chan string) {
	id := fmt.Sprintf("%016x", rand.Uint64())
	fmt.Printf("New Sub: %s\n", id)
	ch := make(chan string, 8)
	b.mu.Lock()
	b.Channels[id] = ch
	b.mu.Unlock()

	return id, ch

}

func (b *Broker) RemoveSubscriber(id string) {
	b.mu.Lock()
	if ch, ok := b.Channels[id]; ok {
		close(ch)
		delete(b.Channels, id)
	}
	b.mu.Unlock()

}

func (b *Broker) Broadcast(msg string) {
	b.mu.RLock()
	for id, ch := range b.Channels {
		select {
		case ch <- msg:
		default:
			fmt.Printf("Broker: channel full for subscriber %s, dropping message\n", id)
		}
	}
	b.mu.RUnlock()
}

func (b *Broker) EventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE Not Supported", http.StatusInternalServerError)
		return
	}

	id, ch := b.AddSubscriber()
	defer b.RemoveSubscriber(id)

	heart := time.NewTicker(15 * time.Second)
	defer heart.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
				fmt.Printf("Broker: write failed for subscriber %s: %v\n", id, err)
				return
			}
			flusher.Flush()
		case <-heart.C:
			data, _ := json.Marshal(map[string]int{"type": 3})
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				fmt.Printf("Broker: heartbeat failed for subscriber %s: %v\n", id, err)
				return
			}
			flusher.Flush()
		}
	}

}
