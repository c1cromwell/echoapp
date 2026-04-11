package infra

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL       string
	ClusterID string
}

// NATSClient wraps nats.go for pub/sub event distribution.
type NATSClient struct {
	conn *nats.Conn
	mu   sync.RWMutex
	subs map[string]*nats.Subscription
}

// NewNATSClient connects to a NATS server.
func NewNATSClient(cfg NATSConfig) (*NATSClient, error) {
	opts := []nats.Option{
		nats.Name("echoapp"),
		nats.Timeout(5 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(60),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Println("NATS reconnected")
		}),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	return &NATSClient{
		conn: conn,
		subs: make(map[string]*nats.Subscription),
	}, nil
}

// Close drains and closes the NATS connection.
func (n *NATSClient) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, sub := range n.subs {
		sub.Unsubscribe()
	}
	n.conn.Close()
	return nil
}

// --- Group Message Fan-Out ---

// GroupMessage is the envelope for fan-out via NATS.
type GroupMessage struct {
	GroupID       string `json:"groupId"`
	MessageID     string `json:"messageId"`
	SenderDID     string `json:"senderDid"`
	EncryptedBlob []byte `json:"encryptedBlob"`
	Timestamp     int64  `json:"timestamp"`
}

// PublishGroupMessage publishes an encrypted group message for fan-out.
func (n *NATSClient) PublishGroupMessage(groupID string, msg GroupMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal group message: %w", err)
	}
	subject := "echo.group." + groupID
	return n.conn.Publish(subject, data)
}

// SubscribeGroupMessages subscribes to messages for a specific group.
func (n *NATSClient) SubscribeGroupMessages(groupID string, handler func(GroupMessage)) error {
	subject := "echo.group." + groupID
	sub, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		var msg GroupMessage
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			log.Printf("unmarshal group message: %v", err)
			return
		}
		handler(msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	n.mu.Lock()
	n.subs[subject] = sub
	n.mu.Unlock()
	return nil
}

// UnsubscribeGroup removes the subscription for a group.
func (n *NATSClient) UnsubscribeGroup(groupID string) error {
	subject := "echo.group." + groupID
	n.mu.Lock()
	sub, ok := n.subs[subject]
	if ok {
		delete(n.subs, subject)
	}
	n.mu.Unlock()
	if ok {
		return sub.Unsubscribe()
	}
	return nil
}

// --- Broadcast Channel Fan-Out ---

// BroadcastPost is the envelope for broadcast channel messages.
type BroadcastPost struct {
	ChannelID     string `json:"channelId"`
	PostID        string `json:"postId"`
	AuthorDID     string `json:"authorDid"`
	EncryptedBlob []byte `json:"encryptedBlob"`
	Timestamp     int64  `json:"timestamp"`
}

// PublishBroadcast publishes a broadcast channel post.
func (n *NATSClient) PublishBroadcast(channelID string, post BroadcastPost) error {
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("marshal broadcast: %w", err)
	}
	return n.conn.Publish("echo.broadcast."+channelID, data)
}

// SubscribeBroadcast subscribes to a broadcast channel.
func (n *NATSClient) SubscribeBroadcast(channelID string, handler func(BroadcastPost)) error {
	subject := "echo.broadcast." + channelID
	sub, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		var post BroadcastPost
		if err := json.Unmarshal(m.Data, &post); err != nil {
			log.Printf("unmarshal broadcast: %v", err)
			return
		}
		handler(post)
	})
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	n.mu.Lock()
	n.subs[subject] = sub
	n.mu.Unlock()
	return nil
}

// --- Generic Pub/Sub ---

// Publish sends a raw message to a subject.
func (n *NATSClient) Publish(subject string, data []byte) error {
	return n.conn.Publish(subject, data)
}

// Subscribe registers a handler for a subject.
func (n *NATSClient) Subscribe(subject string, handler func([]byte)) error {
	sub, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Data)
	})
	if err != nil {
		return err
	}
	n.mu.Lock()
	n.subs[subject] = sub
	n.mu.Unlock()
	return nil
}

// IsConnected returns true if the NATS connection is active.
func (n *NATSClient) IsConnected() bool {
	return n.conn.IsConnected()
}
