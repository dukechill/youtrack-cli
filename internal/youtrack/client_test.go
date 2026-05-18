package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"youtrack-cli/internal/config"
)

func TestSetIssueSprintPostsCommand(t *testing.T) {
	t.Parallel()

	var gotQuery string
	var gotIssueID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/commands" {
			t.Fatalf("expected /api/commands, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		gotQuery, _ = body["query"].(string)
		issues, _ := body["issues"].([]any)
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}

		issue, _ := issues[0].(map[string]any)
		gotIssueID, _ = issue["idReadable"].(string)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.Config{URL: server.URL, Token: "token"}
	if err := SetIssueSprint(cfg, "CT-123", " CRM促案管理 ", " sprint 45 "); err != nil {
		t.Fatalf("SetIssueSprint returned error: %v", err)
	}

	if gotQuery != `board CRM促案管理 sprint 45` {
		t.Fatalf("unexpected query: %q", gotQuery)
	}
	if gotIssueID != "CT-123" {
		t.Fatalf("unexpected issue id: %q", gotIssueID)
	}
}

func TestSetIssuesSprintPostsAllIssueIDs(t *testing.T) {
	t.Parallel()

	var gotIssueIDs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		issues, _ := body["issues"].([]any)
		for _, item := range issues {
			issue, _ := item.(map[string]any)
			gotIssueIDs = append(gotIssueIDs, issue["idReadable"].(string))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.Config{URL: server.URL, Token: "token"}
	if err := SetIssuesSprint(cfg, []string{"CT-123", "CT-456"}, "CRM促案管理", "sprint 45"); err != nil {
		t.Fatalf("SetIssuesSprint returned error: %v", err)
	}

	if len(gotIssueIDs) != 2 || gotIssueIDs[0] != "CT-123" || gotIssueIDs[1] != "CT-456" {
		t.Fatalf("unexpected issue ids: %#v", gotIssueIDs)
	}
}

func TestListIssueSprints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/issues/CT-123/sprints" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		fields := r.URL.Query().Get("fields")
		if fields != "id,name,isCurrent,start,finish,archived" {
			t.Fatalf("unexpected fields: %q", fields)
		}

		_, _ = w.Write([]byte(`[
			{"id":"122-1189","name":"sprint 44","isCurrent":false,"start":1710000000000,"finish":1710600000000,"archived":false},
			{"id":"122-1190","name":"sprint 45","isCurrent":true,"start":1710600000000,"finish":1711200000000,"archived":false}
		]`))
	}))
	defer server.Close()

	cfg := config.Config{URL: server.URL, Token: "token"}
	sprints, err := ListIssueSprints(cfg, "CT-123")
	if err != nil {
		t.Fatalf("ListIssueSprints returned error: %v", err)
	}

	if len(sprints) != 2 {
		t.Fatalf("expected 2 sprints, got %d", len(sprints))
	}
	if sprints[1].Name != "sprint 45" || !sprints[1].IsCurrent {
		t.Fatalf("unexpected sprint result: %+v", sprints[1])
	}
}

func TestInspectIssueFetchesInventoryEvidence(t *testing.T) {
	t.Parallel()

	requested := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		requested[r.URL.Path] = true

		switch r.URL.Path {
		case "/api/issues/CT-123":
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "customFields") {
				t.Fatalf("unexpected fields: %q", fields)
			}
			_, _ = w.Write([]byte(`{
				"idReadable":"CT-123",
				"summary":"Inventory review",
				"customFields":[
					{"name":"State","value":{"name":"In Progress"}},
					{"name":"Priority","value":{"name":"Major"}},
					{"name":"Estimation","value":{"presentation":"2h"}}
				]
			}`))
		case "/api/issues/CT-123/comments":
			_, _ = w.Write([]byte(`[{
				"id":"c1",
				"created":1710000000000,
				"text":"Today: implemented inspect.",
				"author":{"login":"duke","fullName":"Duke Chiu"}
			}]`))
		case "/api/issues/CT-123/timeTracking/workItems":
			_, _ = w.Write([]byte(`[{
				"date":1710003600000,
				"duration":{"minutes":60,"presentation":"1h"},
				"text":"Added tests",
				"author":{"login":"duke","fullName":"Duke Chiu"}
			}]`))
		case "/api/issues/CT-123/sprints":
			_, _ = w.Write([]byte(`[{"id":"s1","name":"sprint 46","isCurrent":true}]`))
		case "/api/issues/CT-123/links":
			_, _ = w.Write([]byte(`[{
				"linkType":{"name":"Depend","localizedName":"depends on"},
				"issues":[{"idReadable":"CT-456","summary":"Dependency"}]
			}]`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := config.Config{URL: server.URL, Token: "token"}
	inspect, err := InspectIssue(cfg, "CT-123")
	if err != nil {
		t.Fatalf("InspectIssue returned error: %v", err)
	}

	if inspect.Issue.ID != "CT-123" || inspect.Issue.Summary != "Inventory review" {
		t.Fatalf("unexpected issue: %+v", inspect.Issue)
	}
	if len(inspect.Comments) != 1 || inspect.Comments[0].Author.FullName != "Duke Chiu" {
		t.Fatalf("unexpected comments: %+v", inspect.Comments)
	}
	if len(inspect.WorkItems) != 1 || inspect.WorkItems[0].Duration.Presentation != "1h" {
		t.Fatalf("unexpected work items: %+v", inspect.WorkItems)
	}
	if len(inspect.Sprints) != 1 || !inspect.Sprints[0].IsCurrent {
		t.Fatalf("unexpected sprints: %+v", inspect.Sprints)
	}
	if len(inspect.Links) != 1 || inspect.Links[0].Issues[0].ID != "CT-456" {
		t.Fatalf("unexpected links: %+v", inspect.Links)
	}
	if len(inspect.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", inspect.Warnings)
	}

	for _, path := range []string{
		"/api/issues/CT-123",
		"/api/issues/CT-123/comments",
		"/api/issues/CT-123/timeTracking/workItems",
		"/api/issues/CT-123/sprints",
		"/api/issues/CT-123/links",
	} {
		if !requested[path] {
			t.Fatalf("expected request to %s", path)
		}
	}
}

func TestDetermineCurrentSprintUsesCurrentFlag(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/agiles":
			_, _ = w.Write([]byte(`[{"id":"121-114","name":"CRM促案管理"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/agiles/121-114/sprints":
			if got := r.URL.Query().Get("fields"); !strings.Contains(got, "isCurrent") {
				t.Fatalf("unexpected fields query: %q", got)
			}
			_, _ = w.Write([]byte(`[
				{"id":"122-1189","name":"sprint 44","isCurrent":false,"start":1710000000000,"finish":1710600000000},
				{"id":"122-1190","name":"sprint 45","isCurrent":true,"start":1710600000000,"finish":1711200000000}
			]`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := config.Config{
		URL:       server.URL,
		Token:     "token",
		BoardName: "CRM促案管理",
	}

	sprintName, err := DetermineCurrentSprint(cfg)
	if err != nil {
		t.Fatalf("DetermineCurrentSprint returned error: %v", err)
	}

	if sprintName != "sprint 45" {
		t.Fatalf("unexpected sprint name: %q", sprintName)
	}
}
