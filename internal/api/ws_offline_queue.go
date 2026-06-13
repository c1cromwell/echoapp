package api

import (
	"sync"
	"time"
)

const (
	wsOfflineMaxDepth    = 1000
	wsOfflineRetention   = 30 * 24 * time.Hour // 1:1 + group directed WS blobs
	wsOfflineGroupRetain = 7 * 24 * time.Hour  // large-group policy hook (same default for M2)
)

type wsOfflineEntry struct {
	data      []byte
	enqueued  time.Time
	retention time.Duration
}

// wsOfflineQueue stores directed WS payloads for offline recipients (M2 group text + group_key).
type wsOfflineQueue struct {
	mu     sync.Mutex
	queues map[string][]wsOfflineEntry
}

func newWSOfflineQueue() *wsOfflineQueue {
	return &wsOfflineQueue{queues: make(map[string][]wsOfflineEntry)}
}

func (q *wsOfflineQueue) Enqueue(recipient string, data []byte, retention time.Duration) {
	if recipient == "" || len(data) == 0 {
		return
	}
	if retention <= 0 {
		retention = wsOfflineRetention
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	queue := q.queues[recipient]
	if len(queue) >= wsOfflineMaxDepth {
		queue = queue[1:]
	}
	queue = append(queue, wsOfflineEntry{
		data:      append([]byte(nil), data...),
		enqueued:  time.Now(),
		retention: retention,
	})
	q.queues[recipient] = queue
}

func (q *wsOfflineQueue) DequeueAll(recipient string) [][]byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	entries := q.queues[recipient]
	delete(q.queues, recipient)
	now := time.Now()
	out := make([][]byte, 0, len(entries))
	for _, e := range entries {
		if now.Sub(e.enqueued) < e.retention {
			out = append(out, e.data)
		}
	}
	return out
}
