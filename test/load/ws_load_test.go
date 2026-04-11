package load

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/testutil"
)

// TestLoad_1000ConcurrentConnections verifies the server can handle
// 1,000 simultaneous WebSocket connections (GO-113).
func TestLoad_1000ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	const numClients = 1000
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	var (
		connected int64
		failed    int64
		wg        sync.WaitGroup
		conns     = make([]*websocket.Conn, numClients)
		mu        sync.Mutex
	)

	// Open 1000 connections concurrently
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(idx int) {
			defer wg.Done()
			token := ts.IssueTestToken(fmt.Sprintf("user-%04d", idx))
			wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws?token=" + token
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}
			atomic.AddInt64(&connected, 1)
			mu.Lock()
			conns[idx] = conn
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	t.Logf("Connected: %d / %d (failed: %d)", connected, numClients, failed)
	if connected < int64(numClients*95/100) {
		t.Fatalf("expected at least 95%% connections, got %d/%d", connected, numClients)
	}

	// Cleanup
	for _, c := range conns {
		if c != nil {
			c.Close()
		}
	}
}

// TestLoad_MessageThroughput measures how many messages per second
// the server can route between connected clients.
func TestLoad_MessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	const (
		numSenders   = 50
		numReceivers = 50
		msgsPerSend  = 20
	)

	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	// Connect receivers first
	receivers := make([]*websocket.Conn, numReceivers)
	for i := 0; i < numReceivers; i++ {
		token := ts.IssueTestToken(fmt.Sprintf("recv-%04d", i))
		wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws?token=" + token
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("receiver %d dial: %v", i, err)
		}
		defer conn.Close()
		receivers[i] = conn
	}

	// Connect senders
	senders := make([]*websocket.Conn, numSenders)
	for i := 0; i < numSenders; i++ {
		token := ts.IssueTestToken(fmt.Sprintf("send-%04d", i))
		wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws?token=" + token
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("sender %d dial: %v", i, err)
		}
		defer conn.Close()
		senders[i] = conn
	}

	time.Sleep(100 * time.Millisecond)

	var received int64
	var recvWg sync.WaitGroup

	// Start receivers reading in background
	for i := 0; i < numReceivers; i++ {
		recvWg.Add(1)
		go func(conn *websocket.Conn) {
			defer recvWg.Done()
			for {
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
				atomic.AddInt64(&received, 1)
			}
		}(receivers[i])
	}

	// All senders send concurrently
	start := time.Now()
	var sendWg sync.WaitGroup
	var sent int64

	for i := 0; i < numSenders; i++ {
		sendWg.Add(1)
		go func(senderIdx int, conn *websocket.Conn) {
			defer sendWg.Done()
			targetRecv := fmt.Sprintf("recv-%04d", senderIdx%numReceivers)
			for j := 0; j < msgsPerSend; j++ {
				msg := api.WSMessage{
					Type:    "text",
					To:      targetRecv,
					Payload: json.RawMessage(fmt.Sprintf(`"msg-%d-%d"`, senderIdx, j)),
				}
				data, _ := json.Marshal(msg)
				if err := conn.WriteMessage(websocket.TextMessage, data); err == nil {
					atomic.AddInt64(&sent, 1)
				}
			}
		}(i, senders[i])
	}
	sendWg.Wait()
	elapsed := time.Since(start)

	// Wait for receivers to drain
	time.Sleep(2 * time.Second)
	for _, c := range receivers {
		c.Close()
	}
	recvWg.Wait()

	totalSent := atomic.LoadInt64(&sent)
	totalRecv := atomic.LoadInt64(&received)
	throughput := float64(totalSent) / elapsed.Seconds()

	t.Logf("Sent: %d messages in %v (%.0f msg/s)", totalSent, elapsed, throughput)
	t.Logf("Received: %d / %d (%.1f%% delivery)", totalRecv, totalSent, float64(totalRecv)/float64(totalSent)*100)

	// At minimum, expect >90% delivery rate
	if totalRecv < totalSent*90/100 {
		t.Errorf("delivery rate below 90%%: %d/%d", totalRecv, totalSent)
	}
}

// TestLoad_ConnectionLatency measures individual WebSocket connection time.
func TestLoad_ConnectionLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	const trials = 100
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	var totalDuration time.Duration
	for i := 0; i < trials; i++ {
		token := ts.IssueTestToken(fmt.Sprintf("latency-%04d", i))
		wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws?token=" + token

		start := time.Now()
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("trial %d: %v", i, err)
		}
		conn.Close()
		totalDuration += elapsed
	}

	avg := totalDuration / trials
	t.Logf("Avg connection latency over %d trials: %v", trials, avg)
	if avg > 100*time.Millisecond {
		t.Errorf("average connection latency too high: %v (want < 100ms)", avg)
	}
}
