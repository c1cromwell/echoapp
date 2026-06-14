package api

import (
	"context"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

const (
	wsOfflineMaxDepth    = 1000
	wsOfflineRetention   = 30 * 24 * time.Hour // 1:1 + group directed WS blobs
	wsOfflineGroupRetain = 7 * 24 * time.Hour  // large-group policy hook (same default for M2)
)

type wsOfflineEntry struct {
	data       []byte
	storageURI string // WO-237: encblob URI when queue depth exceeded
	enqueued   time.Time
	retention  time.Duration
}

// wsOfflineDrain is the result of draining a recipient's offline queue on reconnect.
type wsOfflineDrain struct {
	Blobs        [][]byte
	OverflowURIs []string
}

// wsOfflineQueue stores directed WS payloads for offline recipients (M2 group text + group_key).
type wsOfflineQueue struct {
	mu       sync.Mutex
	queues   map[string][]wsOfflineEntry
	overflow encblob.Storage
}

func newWSOfflineQueue() *wsOfflineQueue {
	return &wsOfflineQueue{queues: make(map[string][]wsOfflineEntry)}
}

func (q *wsOfflineQueue) SetOverflowStorage(s encblob.Storage) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.overflow = s
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
	now := time.Now()

	if len(queue) >= wsOfflineMaxDepth {
		// WO-237: pin overflow blobs to content-blind storage; queue only the URI.
		if q.overflow != nil {
			if uri, err := q.overflow.Store(context.Background(), data); err == nil && uri != "" {
				queue = append(queue, wsOfflineEntry{
					storageURI: uri,
					enqueued:   now,
					retention:  retention,
				})
				q.queues[recipient] = queue
			}
		}
		return
	}

	queue = append(queue, wsOfflineEntry{
		data:      append([]byte(nil), data...),
		enqueued:  now,
		retention: retention,
	})
	q.queues[recipient] = queue
}

func (q *wsOfflineQueue) DequeueAll(recipient string) wsOfflineDrain {
	q.mu.Lock()
	defer q.mu.Unlock()
	entries := q.queues[recipient]
	delete(q.queues, recipient)
	now := time.Now()
	out := wsOfflineDrain{}
	for _, e := range entries {
		if now.Sub(e.enqueued) >= e.retention {
			continue
		}
		if e.storageURI != "" {
			out.OverflowURIs = append(out.OverflowURIs, e.storageURI)
			continue
		}
		out.Blobs = append(out.Blobs, e.data)
	}
	return out
}
