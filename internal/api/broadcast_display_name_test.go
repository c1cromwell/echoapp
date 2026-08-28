package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

func TestBroadcastMembers_IncludesDisplayName(t *testing.T) {
	db := database.NewMemoryDB()
	ctx := context.Background()
	if err := db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:key:alice", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateUser(ctx, &database.User{UserID: "u2", DID: "did:key:bob", Username: "bobby"}); err != nil {
		t.Fatal(err)
	}
	csvc := contacts.NewService(db)
	name := "Bob Rivera"
	if _, err := csvc.UpdateOwnProfile(ctx, "did:key:bob", &name, nil, nil, nil); err != nil {
		t.Fatal(err)
	}

	bcast := broadcast_channels.NewChannelService()
	ch, err := bcast.CreateChannel("Town", "news", "did:key:alice", broadcast_channels.ChannelTypeCommunity)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bcast.Subscribe(ch.ID, "did:key:bob"); err != nil {
		t.Fatal(err)
	}

	h := &V3Handlers{DB: db, Contacts: csvc, Broadcasts: bcast}
	req := httptest.NewRequest(http.MethodGet, "/v3/broadcasts/members?channelId="+ch.ID, nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:alice"))
	rec := httptest.NewRecorder()
	h.handleBroadcastMembers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Members []struct {
			SubscriberID string `json:"subscriber_id"`
			DisplayName  string `json:"display_name"`
		} `json:"members"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range body.Members {
		if m.SubscriberID == "did:key:bob" {
			found = true
			if m.DisplayName != "Bob Rivera" {
				t.Fatalf("bob display_name = %q", m.DisplayName)
			}
		}
	}
	if !found {
		t.Fatal("bob missing from members")
	}
}

func TestBroadcastComments_IncludesAuthorDisplayName(t *testing.T) {
	db := database.NewMemoryDB()
	ctx := context.Background()
	if err := db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:key:alice", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	csvc := contacts.NewService(db)
	name := "Alice Chen"
	if _, err := csvc.UpdateOwnProfile(ctx, "did:key:alice", &name, nil, nil, nil); err != nil {
		t.Fatal(err)
	}

	bcast := broadcast_channels.NewChannelService()
	ch, err := bcast.CreateChannel("Town", "news", "did:key:alice", broadcast_channels.ChannelTypeCommunity)
	if err != nil {
		t.Fatal(err)
	}
	post, err := bcast.CreatePost(ch.ID, "did:key:alice", "hello", broadcast_channels.ContentTypeText)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bcast.AddComment(ch.ID, post.ID, "did:key:alice", "nice"); err != nil {
		t.Fatal(err)
	}

	h := &V3Handlers{DB: db, Contacts: csvc, Broadcasts: bcast}
	req := httptest.NewRequest(http.MethodGet, "/v3/broadcasts/comments?postId="+post.ID, nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:alice"))
	rec := httptest.NewRecorder()
	h.handleBroadcastComments(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Comments []struct {
			AuthorID    string `json:"author_id"`
			DisplayName string `json:"display_name"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Comments) != 1 || body.Comments[0].DisplayName != "Alice Chen" {
		t.Fatalf("comments = %+v", body.Comments)
	}
}
