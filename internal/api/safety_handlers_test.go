package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

const testAdminToken = "test-admin-token-0123456789abcdef0123"

func adminGet(rt *Router, target, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	return rec
}

// A report can say which message / conversation / channel it is about and
// carry bounded free text, and the free text is not echoed to the reporter.
func TestSafetyReport_StoresReferencesAndDetails(t *testing.T) {
	rt := NewRouter([]string{"*"})
	db := database.NewMemoryDB()
	rt.V3 = &V3Handlers{DB: db}
	token := accessTokenFor(t, rt, "did:key:alice")

	rec := secPostJSON(t, rt, "/v1/safety/reports", map[string]string{
		"subject_did":     "did:key:bob",
		"reason":          "harassment",
		"context":         "message",
		"message_id":      "msg-42",
		"conversation_id": "conv-7",
		"details":         "  keeps sending threats  ",
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d %s", rec.Code, rec.Body.String())
	}
	stored, _ := db.ListSafetyReportsByReporter(context.Background(), "did:key:alice")
	if len(stored) != 1 {
		t.Fatalf("stored = %+v", stored)
	}
	got := stored[0]
	if got.MessageID != "msg-42" || got.ConversationID != "conv-7" || got.ChannelID != "" || got.Details != "keeps sending threats" {
		t.Fatalf("stored report = %+v", got)
	}

	// The reporter's own list does not return the free text.
	req := httptest.NewRequest(http.MethodGet, "/v1/safety/reports", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	list := httptest.NewRecorder()
	rt.Handler().ServeHTTP(list, req)
	if list.Code != http.StatusOK {
		t.Fatalf("list: %d %s", list.Code, list.Body.String())
	}
	if strings.Contains(list.Body.String(), "threats") || !strings.Contains(list.Body.String(), "msg-42") {
		t.Fatalf("reporter list should carry the references but not the details: %s", list.Body.String())
	}
}

func TestSafetyReport_BoundsUserText(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.V3 = &V3Handlers{DB: database.NewMemoryDB()}
	token := accessTokenFor(t, rt, "did:key:alice")
	base := map[string]string{"subject_did": "did:key:bob", "reason": "spam", "context": "channel"}

	withField := func(key, value string) map[string]string {
		out := map[string]string{key: value}
		for k, v := range base {
			out[k] = v
		}
		return out
	}
	// Characters, not bytes: 1000 multi-byte runes are still within the cap.
	if rec := secPostJSON(t, rt, "/v1/safety/reports", withField("details", strings.Repeat("é", database.SafetyReportMaxDetails)), token); rec.Code != http.StatusCreated {
		t.Fatalf("details at the cap: want 201, got %d %s", rec.Code, rec.Body.String())
	}
	if rec := secPostJSON(t, rt, "/v1/safety/reports", withField("details", strings.Repeat("a", database.SafetyReportMaxDetails+1)), token); rec.Code != http.StatusBadRequest {
		t.Fatalf("details over the cap: want 400, got %d %s", rec.Code, rec.Body.String())
	}
	if rec := secPostJSON(t, rt, "/v1/safety/reports", withField("channel_id", strings.Repeat("c", safetyReportMaxIDLen+1)), token); rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized channel_id: want 400, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSafetyReports_OperatorOnly(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.V3 = &V3Handlers{DB: database.NewMemoryDB()}
	userToken := accessTokenFor(t, rt, "did:key:alice")

	// No operator token configured: the route does not exist, even for a
	// signed-in user.
	t.Setenv(adminAPITokenEnv, "")
	if rec := adminGet(rt, adminSafetyReportsPath, userToken); rec.Code != http.StatusNotFound {
		t.Fatalf("unset token: want 404, got %d %s", rec.Code, rec.Body.String())
	}
	// A token too short to be a secret counts as unset.
	t.Setenv(adminAPITokenEnv, "short")
	if rec := adminGet(rt, adminSafetyReportsPath, "short"); rec.Code != http.StatusNotFound {
		t.Fatalf("weak token: want 404, got %d %s", rec.Code, rec.Body.String())
	}

	t.Setenv(adminAPITokenEnv, testAdminToken)
	for name, bearer := range map[string]string{
		"no credentials":      "",
		"a user access token": userToken,
		"wrong token":         testAdminToken + "x",
	} {
		if rec := adminGet(rt, adminSafetyReportsPath, bearer); rec.Code != http.StatusForbidden {
			t.Fatalf("%s: want 403, got %d %s", name, rec.Code, rec.Body.String())
		}
	}
	if rec := adminGet(rt, adminSafetyReportsPath, testAdminToken); rec.Code != http.StatusOK {
		t.Fatalf("operator: want 200, got %d %s", rec.Code, rec.Body.String())
	}
	// The operator token is not a user session anywhere else.
	if rec := adminGet(rt, "/v1/safety/reports", testAdminToken); rec.Code != http.StatusUnauthorized {
		t.Fatalf("operator token on a user route: want 401, got %d %s", rec.Code, rec.Body.String())
	}
}

func testAdminSafetyReportsListing(t *testing.T, rt *Router) {
	t.Helper()
	t.Setenv(adminAPITokenEnv, testAdminToken)
	alice := accessTokenFor(t, rt, "did:key:alice")
	for i := 0; i < 5; i++ {
		subject := "did:key:bob"
		if i == 2 {
			subject = "did:key:carol"
		}
		rec := secPostJSON(t, rt, "/v1/safety/reports", map[string]string{
			"subject_did": subject,
			"reason":      "spam",
			"context":     "message",
			"message_id":  fmt.Sprintf("m%d", i),
			"details":     fmt.Sprintf("detail %d", i),
		}, alice)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}

	type page struct {
		Reports    []database.SafetyReport `json:"reports"`
		NextOffset *int                    `json:"next_offset"`
	}
	fetch := func(query string) page {
		rec := adminGet(rt, adminSafetyReportsPath+query, testAdminToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", query, rec.Code, rec.Body.String())
		}
		var out page
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}

	// By subject, newest first, with the details and the reference.
	bob := fetch("?subject_did=did:key:bob&limit=3")
	if len(bob.Reports) != 3 || bob.NextOffset == nil || *bob.NextOffset != 3 {
		t.Fatalf("first page = %+v", bob)
	}
	if bob.Reports[0].MessageID != "m4" || bob.Reports[0].Details != "detail 4" || bob.Reports[0].ReporterDID != "did:key:alice" {
		t.Fatalf("newest report = %+v", bob.Reports[0])
	}
	for _, report := range bob.Reports {
		if report.SubjectDID != "did:key:bob" {
			t.Fatalf("subject filter leaked %+v", report)
		}
	}
	rest := fetch("?subject_did=did:key:bob&limit=3&offset=3")
	if len(rest.Reports) != 1 || rest.NextOffset != nil || rest.Reports[0].MessageID != "m0" {
		t.Fatalf("second page = %+v", rest)
	}
	if all := fetch(""); len(all.Reports) != 5 {
		t.Fatalf("unfiltered = %d reports", len(all.Reports))
	}
	if rec := adminGet(rt, adminSafetyReportsPath+"?limit=0", testAdminToken); rec.Code != http.StatusBadRequest {
		t.Fatalf("limit=0: want 400, got %d", rec.Code)
	}
}

func TestAdminSafetyReports_ListsBySubjectWithPagination(t *testing.T) {
	t.Run("database", func(t *testing.T) {
		rt := NewRouter([]string{"*"})
		rt.V3 = &V3Handlers{DB: database.NewMemoryDB()}
		testAdminSafetyReportsListing(t, rt)
	})
	t.Run("router fallback store", func(t *testing.T) {
		testAdminSafetyReportsListing(t, NewRouter([]string{"*"}))
	})
}
